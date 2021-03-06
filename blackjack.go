package main

import (
	"fmt"
	"github.com/hoisie/web"
	"github.com/kr/pretty"
	"math/rand"
	"time"
)

//
//   CONSTANTS
//
const (
	HouseMinFunds        = 5
	HouseMinBet          = 5
	HouseDealerStay      = 17
	HousePayoffFactor    = 1.0
	HouseBlackjackFactor = 1.5
	HouseDfltDeckCount   = 2
	HouseFunds           = 1000000
	BaseDeckLen          = 52
	BlackjackScore       = 21
	MinDeckSum           = 2*BlackjackScore + 2
	NoBetYet             = "You haven't placed a bet yet!\n\n"
	NoDealYet            = "No opening deal yet!\n\n"
	DealerShowing        = "Dealer showing: %# v\nYou have: %# v\nSum: %d\n\n"
	YouWin               = "You win %.2f parsohns of space cash\nYou: %d\nDealer: %d\n"
	YouLose              = "You lose %.2f parsohns of space cash\nYou: %d\nDealer: %d\n"
	YouBusted            = "You are bust, loss of %.2f parsohns\n\n"
)
const (
	Win = iota
	Loss
	Push
	Blackjack
)

// command tags for passing to dispatch
const (
	DealCmd = iota
	HitCmd
	StayCmd
	BetCmd
	HandCmd
	DepositCmd
	FundsCmd
	ShowDeckCmd
	ResizeDeckCmd
)

//
//   data type declarations
//

type Card struct {
	value int
	name  string
}

type Deck []Card
type Hand []Card

type Player struct {
	name  string
	hand  Hand
	funds float64
	bet   float64
}

type Dealer struct {
	hand  Hand
	funds float64
}

type Cmd struct {
	code  int
	value string
}

type State struct {
	dealer    Dealer
	me        Player
	deckCount int
	deck      Deck
	cmdChan   chan Cmd
}

//
//  GLOBAL DATA
//

// our template deck - aces defaulted to value 11
var fullDeck = Deck{
	{2, "2C"}, {3, "3C"}, {4, "4C"}, {5, "5C"}, {6, "6C"}, {7, "7C"}, {8, "8C"},
	{9, "9C"}, {10, "10C"}, {10, "JC"}, {10, "QC"}, {10, "KC"}, {11, "AC"},
	{2, "2D"}, {3, "3D"}, {4, "4D"}, {5, "5D"}, {6, "6D"}, {7, "7D"}, {8, "8D"},
	{9, "9D"}, {10, "10D"}, {10, "JD"}, {10, "QD"}, {10, "KD"}, {11, "AD"},
	{2, "2H"}, {3, "3H"}, {4, "4H"}, {5, "5H"}, {6, "6H"}, {7, "7H"}, {8, "8H"},
	{9, "9H"}, {10, "10H"}, {10, "JH"}, {10, "QH"}, {10, "KH"}, {11, "AH"},
	{2, "2S"}, {3, "3S"}, {4, "4S"}, {5, "5S"}, {6, "6S"}, {7, "7S"}, {8, "8S"},
	{9, "9S"}, {10, "10S"}, {10, "JS"}, {10, "QS"}, {10, "KS"}, {11, "AS"},
}

var playerList = make(map[string]State)

//
//  load the callbacks and start the http server
//
func main() {

	// used in shuffle algorithm
	rand.Seed(time.Now().UnixNano())

	// player api
	web.Get("/deal", deal_cmd)
	web.Get("/hit", hit_cmd)
	web.Get("/stay", stay_cmd)
	web.Get("/bet", bet_cmd)
	web.Get("/hand", hand_cmd)
	web.Get("/deposit", deposit_cmd)
	web.Get("/funds", funds_cmd)

	// admin api
	web.Post("/show_deck", deck_show_cmd)
	web.Post("/resize_deck", deck_resize_cmd)

	// setup api
	web.Get("/create_player", create_player)
	web.Get("/create_game", create_game)

	// kick off the webserver
	web.Run("0.0.0.0:9999")

}

//
//  PUBLIC API CALLBACKS
//

// curl http://localhost:9999/deal?name=NAME
func deal_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{DealCmd, ""}

}

// curl http://localhost:9999/hit?name=NAME
func hit_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{HitCmd, ""}

}

// curl http://localhost:9999/stay?name=NAME
func stay_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{StayCmd, ""}

}

// curl http://localhost:9999/bet?name=NAME[&amount=XXX]
func bet_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{BetCmd, ctx.Params["amount"]}

}

// curl http://localhost:9999/hand?name=NAME
func hand_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{HandCmd, ""}

}

// curl http://localhost:9999/deposit?name=NAME[&amount=XXX]
func deposit_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{DepositCmd, ctx.Params["amount"]}

}

// curl http://localhost:9999/funds?name=NAME
func funds_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{FundsCmd, ""}

}

// curl --data auth=titanoboa http://localhost:9999/show_deck?name=NAME
func deck_show_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{ShowDeckCmd, ctx.Params["auth"]}

}

// curl http://localhost:9999/deck_resize?name=NAME&count=XXX
func deck_resize_cmd(ctx *web.Context) {

	playerList[ctx.Params["name"]].cmdChan <- Cmd{ResizeDeckCmd, ctx.Params["count"]}

}

// curl http://localhost:9999/create_player?name=NAME&amount=XXX
func create_player(ctx *web.Context) {

	// add the player NAME to the list of players if new
	var startFunds float64
	fmt.Sscanf(ctx.Params["amount"], "%g", &startFunds)
	name := ctx.Params["name"]

	// create State object, allocate comms channels, init bank funds
	if _, ok := playerList[name]; !ok {
		playerList[name] = State{
			Dealer{nil, HouseFunds},
			Player{name, nil, startFunds, 0},
			HouseDfltDeckCount, nil, nil,
		}
		fmt.Printf("Player created: %# v\n\n", playerList[name])
	}

	return

}

// curl http://localhost:9999/create_game?name=NAME
func create_game(ctx *web.Context) string {

	var userName string
	fmt.Sscanf(ctx.Params["name"], "%s", &userName)
	if _, ok := playerList[userName]; !ok {
		return fmt.Sprintf("Player %s does not exist - create player first\n\n", userName)
	}

	// update the state object for this new game
	state := playerList[userName]
	state.cmdChan = make(chan Cmd, 1)
	reload(state.deck, state.deckCount)
	playerList[userName] = state

	// the idea here is we encapsulate all game state in the following closure
	go func(name string) {

		for {

			switch cmd := <-state.cmdChan; cmd.code {
			case DealCmd:
				(&state).deal()
			case HitCmd:
				(&state).hit()
			case StayCmd:
				(&state).stay()
			case BetCmd:
				var bet float64
				fmt.Sscanf(cmd.value, "%f", &bet)
				(&state).bet(bet)
			case DepositCmd:
				var deposit float64
				fmt.Sscanf(cmd.value, "%f", &deposit)
				(&state).deposit(deposit)
			case HandCmd:
				(&state).hand()
			case FundsCmd:
				(&state).funds()
			case ShowDeckCmd:
				if cmd.value != "titanoboa" {
					fmt.Print("Incorrect auth\n\n")
					return
				}
				(&state).deck_show()
			case ResizeDeckCmd:
				var count int
				fmt.Sscanf(cmd.value, "%d", &count)
				(&state).deck_resize(count)
			default:
				fmt.Printf("Unknown cmd %# v\n\n", cmd)
				return
			}

			playerList[userName] = state

		}

	}(userName)

	return fmt.Sprintf("Game created, user %s\n\n", userName)
}

func (state *State) deal() {

	if state.me.bet == 0 {
		fmt.Print(NoBetYet)
		return
	}

	if state.deck == nil || state.deck.sum() < MinDeckSum {
		state.deck = reload(state.deck, state.deckCount)
	}

	state.dealer.hand = nil
	state.me.hand = nil

	// NB: hardcoded rule, implicit ordering! weak of me.
	dlr1 := state.deck.pop()
	me1 := state.deck.pop()
	dlr2 := state.deck.pop()
	me2 := state.deck.pop()

	state.dealer.hand = append(state.dealer.hand, dlr1, dlr2)
	state.me.hand = append(state.me.hand, me1, me2)

	if state.me.hand.sum() == BlackjackScore {
		win := state.settle(Blackjack)
		fmt.Printf(YouWin, win, BlackjackScore, dlr1)
		return
	}

	pretty.Printf(DealerShowing, dlr1, state.me.hand, state.me.hand.sum())
	return

}

func (state *State) hit() {

	if state.me.hand == nil {
		fmt.Print(NoDealYet)
		return
	}

	if state.me.bet == 0 {
		fmt.Print(NoBetYet)
		return
	}

	card := state.deck.pop()
	state.me.hand = append(state.me.hand, card)

	sum := state.me.hand.sum()
	if sum > BlackjackScore {
		loss := state.settle(Loss)
		fmt.Printf(YouBusted, loss)
		return
	}

	pretty.Printf(DealerShowing, state.dealer.hand[0], state.me.hand, sum)
	return

}

func (state *State) stay() {

	if state.me.bet == 0 {
		fmt.Print(NoBetYet)
		return
	}

	state.dealer_wrap()

	myScore := state.me.hand.sum()
	dealerScore := state.dealer.hand.sum()
	switch {
	case dealerScore > BlackjackScore || myScore > dealerScore:
		win := state.settle(Win)
		fmt.Printf(YouWin, win, myScore, dealerScore)
	case myScore == dealerScore:
		state.settle(Push)
		fmt.Printf("Push!\n\n")
	default:
		loss := state.settle(Loss)
		fmt.Printf(YouLose, loss, myScore, dealerScore)
	}

	return

}

func (state *State) bet(bet float64) {

	if bet < HouseMinBet {
		fmt.Printf("Bet %.2f is under house minimum (%.2f)\n\n", HouseMinBet, bet)
		return
	}
	if bet > state.me.funds {
		fmt.Printf("Bet %.2f is above your available funds (%.2f)\n\n", bet, state.me.funds)
		return
	}

	state.me.bet += bet
	state.me.funds -= bet

	fmt.Printf("Your current bet: %.2f parsohns of space cash\n\n", state.me.bet)
	return

}

func (state *State) deposit(amount float64) {

	state.me.funds += amount

	fmt.Printf("Deposited %.2f, balance now %.2f\n\n", amount, state.me.funds)
	return

}

func (state *State) hand() {

	if state.me.hand == nil {
		fmt.Print(NoDealYet)
		return
	}

	pretty.Printf(DealerShowing, state.dealer.hand[0], state.me.hand, state.me.hand.sum())
	return

}

func (state *State) funds() {

	fmt.Printf("Your remaining funds: %.2f parsohns of space cash\n\n", state.me.funds)
	return

}

//
//  ADMIN API
//

func (state *State) deck_show() {

	pretty.Printf("Current state of deck: %# v\n\n", state.deck)
	return

}

func (state *State) deck_resize(deckCount int) {

	// reallocate deck at new size
	state.deck = reload(state.deck, deckCount)

	pretty.Printf("New state of deck: %# v\n\n", state.deck)
	return

}

//
//  'State' MAINTENANCE FUNCTIONS
//

// sort out the funds transfers and clear per-game state (hands and bets)
func (state *State) settle(result int) float64 {

	net := 0.0
	switch result {
	case Win:
		net = state.me.bet * HousePayoffFactor
		state.me.funds += (state.me.bet + net)
		state.dealer.funds -= net
	case Blackjack:
		net = state.me.bet * HouseBlackjackFactor
		state.me.funds += (state.me.bet + net)
		state.dealer.funds -= net
	case Loss:
		net = state.me.bet
		state.dealer.funds += net
	case Push:
		net = 0.0
		state.me.funds += state.me.bet
	}

	state.dealer.hand = nil
	state.me.hand = nil
	state.me.bet = 0.0
	return net

}

// after player/s have stayed or busted, this is called to complete the dealer's hand
func (state *State) dealer_wrap() {

	// implement the "hit 'till 17 or bust" rule if necessary
	for state.dealer.hand.sum() < HouseDealerStay {
		card := state.deck.pop()
		state.dealer.hand = append(state.dealer.hand, card)
	}

	return

}

//
//  UTILITY FUNCTIONS
//

// reinitialize deck/s with deckCount fresh packs
func reload(deck Deck, deckCount int) Deck {

	// only reallocate if deck size actually changes
	deckLen := deckCount * BaseDeckLen
	if deck == nil || deckLen != len(deck) {
		deck = nil
		deck = make(Deck, deckLen)
	}

	// init by cut and paste using our template deck
	for i := 0; i < deckLen; i += BaseDeckLen {
		copy(deck[i:], fullDeck)
	}

	deck.shuffle()

	return deck

}

// implement Algorithm P, Knuth/Durstenfeld
func (deck *Deck) shuffle() {

	// TL/DR: swap every element with a random earlier element
	d := *deck
	l := len(d)
	for i := l - 1; i > 0; i-- {
		j := rand.Intn(l - i)
		tmp := d[j]
		d[j] = d[i]
		d[i] = tmp
	}
	*deck = d
	return

}

// pull the top card and reset head of list
func (deck *Deck) pop() Card {

	d := *deck
	card := d[0]
	*deck = d[1:]
	return card

}

// add up this hand strategically
func (hand Hand) sum() int {

	// swap one 11 for a 1 if we're going to bust
	sum, aces := sum_cards(hand)

	if aces > 0 && sum > Blackjack {
		sum -= 10
		fmt.Printf("%d aces! we should split.../n/n", aces)
	}

	return sum

}

// count aces as 1 to produce conservative guess at possible hands
func (deck Deck) sum() int {

	sum, aces := sum_cards(deck)
	return sum - aces*10

}

// add up
func sum_cards(cards []Card) (int, int) {

	aces, sum := 0, 0
	for i := 0; i < len(cards); i++ {
		// count the aces while summing
		if cards[i].value == 11 {
			aces++
		}
		sum += cards[i].value
	}

	return sum, aces

}
