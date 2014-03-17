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
const HouseMinFunds = 5
const HouseMinBet = 5
const HouseDealerStay = 17
const HousePayoffFactor = 2.0    // bet + 100%
const HouseBlackjackFactor = 2.5 // bet + 150%
const HouseDfltDeckCount = 2
const BaseDeckLen = 52
const BlackjackScore = 21
const (
	Win = iota
	Loss
	Push
	Blackjack
)

//
//   data type declarations
//

type Card struct {
	value int
	name  string
}
type Hand []Card

type Player struct {
	hand  Hand
	funds float64
	bet   float64
}
type Dealer struct {
	hand  Hand
	funds float64
}

type Deck []Card

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
var deck Deck
var deckCount = HouseDfltDeckCount

var dealer = Dealer{nil, 1000000}
var me = Player{nil, 0, 0}

//
//  load the callbacks and start the http server
//
func main() {
	// used in shuffle algorithm
	rand.Seed(time.Now().UnixNano())
	// player api
	web.Get("/deal", deal)
	web.Get("/hit", hit)
	web.Get("/stay", stay)
	web.Get("/bet", bet)
	web.Get("/hand", hand)
	web.Get("/deposit", deposit)
	web.Get("/funds", funds)
	// admin api
	web.Post("/show_deck", show_deck)
	web.Post("/size_deck", size_deck)
	// kick it off
	web.Run("0.0.0.0:9999")
}

//
//  PUBLIC API CALLBACKS
//

// curl http://localhost:9999/deal[?reload=y]
func deal(ctx *web.Context) string {

	if me.bet == 0 {
		return fmt.Sprintln("You haven't placed a bet yet!\n")
	}

	if deck == nil || ctx.Params["reload"] == "y" {
		reload()
	}

	dealer.hand = nil
	me.hand = nil

	// NB: hardcoded rule, implicit ordering! weak of me.
	dlr1 := pop()
	me1 := pop()
	dlr2 := pop()
	me2 := pop()

	dealer.hand = append(dealer.hand, dlr1, dlr2)
	me.hand = append(me.hand, me1, me2)

	if me.hand.sum() == BlackjackScore {
		win := settle(Blackjack)
		return fmt.Sprintln("You win %.2f parsohns with a blackjack!\n", win)
	}

	return pretty.Sprintf("Dealer showing: %# v\nYou have: %# v\nSum: %d\n\n",
		dealer.hand[0], me.hand, me.hand.sum())

}

// curl http://localhost:9999/hit
func hit() string {

	if me.hand == nil {
		return fmt.Sprintln("No hits before the opening deal!\n")
	}

	if me.bet == 0 {
		return fmt.Sprintln("You haven't placed a bet yet!\n")
	}

	card := pop()
	me.hand = append(me.hand, card)

	sum := me.hand.sum()
	if sum > BlackjackScore {
		loss := settle(Loss)
		return fmt.Sprintln("You are bust, loss of %.2f parsohns\n", loss)
	}

	return pretty.Sprintf("Dealer showing: %# v\nYou have: %# v\nSum: %d\n\n",
		dealer.hand[0], me.hand, me.hand.sum())

}

// curl http://localhost:9999/hit
func stay() string {

	if me.bet == 0 {
		return fmt.Sprint("You haven't placed a bet yet!\n")
	}

	// the dealer now finishes his hand
	dealer_wrap()

	myScore := me.hand.sum()
	dealerScore := dealer.hand.sum()
	switch {
	case dealerScore > BlackjackScore || myScore > dealerScore:
		win := settle(Win)
		return fmt.Sprintf("You win %.2f parsohns of space cash\n\n", win)
	case myScore == dealerScore:
		settle(Push)
		return fmt.Sprintf("Push!\n\n")
	default:
		loss := settle(Loss)
		return fmt.Sprintf("You lose %.2f parsohns: %d to dealer's %d\n\n",
			loss, myScore, dealerScore)
	}

}

// curl http://localhost:9999/bet?amount=XXX
func bet(ctx *web.Context) string {
	var bet float64

	num, _ := fmt.Sscanf(ctx.Params["amount"], "%f", &bet)
	if num == 0 {
		return fmt.Sprintf("Your current bet: %.2f\n\n", me.bet)
	}

	if bet < HouseMinBet {
		return fmt.Sprintf("Bet %.2f is under house minimum (%.2f)\n\n",
			HouseMinBet, bet)
	}
	if bet > me.funds {
		return fmt.Sprintf("Bet %.2f is above your available funds (%.2f)\n\n",
			bet, me.funds)
	}

	me.bet += bet
	me.funds -= bet

	return fmt.Sprintf("Your bet: %.2f parsohns of space cash\n\n", bet)

}

// curl http://localhost:9999/deposit[?amount=XXX]
func deposit(ctx *web.Context) string {
	var deposit float64

	num, err := fmt.Sscanf(ctx.Params["amount"], "%f", &deposit)
	if num != 1 {
		return fmt.Sprintf("Deposit failed, %# v\n\n", err)
	}

	me.funds += deposit

	return fmt.Sprintf("Deposited %.2f\n\n", deposit)

}

// curl http://localhost:9999/hand
func hand() string {

	if me.hand == nil {
		return fmt.Sprintln("No hand before the opening deal!\n")
	}

	return pretty.Sprintf("Dealer showing: %# v\nYou have: %# v\nSum: %d\n\n",
		dealer.hand[0], me.hand, me.hand.sum())

}

// curl http://localhost:9999/funds
func funds() string {

	return fmt.Sprintf("Your remaining funds: %.2f parsohns of space cash\n\n",
		me.funds)

}

//
//  ADMIN API
//

// curl --data auth=titanoboa http://localhost:9999/show_deck
func show_deck(ctx *web.Context) string {

	// check our "security" - not security
	if ctx.Params["auth"] != "titanoboa" {
		return fmt.Sprintf("Incorrect auth\n\n")
	}

	return pretty.Sprintf("Current state of deck: %# v\n\n", deck)

}

// curl --data auth=titanoboa --data count=XXX http://localhost:9999/size_deck
func size_deck(ctx *web.Context) string {

	if ctx.Params["auth"] != "titanoboa" {
		return fmt.Sprintf("Incorrect auth\n\n")
	}

	num, err := fmt.Sscanf(ctx.Params["count"], "%d", &deckCount)
	if num != 1 {
		return fmt.Sprintf("Deck resize failed, %# v\n\n", err)
	}

	// reallocate deck at new size
	reload()

	return pretty.Sprintf("New state of deck: %# v\n\n", deck)

}

//
//  UTILITY FUNCTIONS
//

// reinitialize deck/s with deckCount fresh packs
func reload() {

	deck = nil
	deckLen := deckCount * BaseDeckLen
	deck = make(Deck, deckLen)

	// init by cut and paste using our template deck
	for i := 0; i < deckLen; i += BaseDeckLen {
		copy(deck[i:], fullDeck)
	}

	deck.shuffle()

	return

}

// implement Algorithm P, Knuth/Durstenfeld
func (deck Deck) shuffle() Deck {

	// TL/DR: swap every element with a random earlier element
	l := len(deck)
	for i := l - 1; i > 0; i-- {
		j := rand.Intn(l - i)
		tmp := deck[j]
		deck[j] = deck[i]
		deck[i] = tmp
	}
	return deck

}

// pull the top card and reset head of list
func pop() Card {

	card := deck[0]
	deck = deck[1:]
	return card

}

// add up this hand
func (hand Hand) sum() int {

	ace, sum := 0, 0
	for i := 0; i < len(hand); i++ {
		// count the aces while summing
		if hand[i].value == 11 {
			ace++
		}
		sum += hand[i].value
	}

	// swap an 11 for a 1 if we're going to bust
	if ace > 0 && sum > Blackjack {
		sum -= 10
		if ace == 2 {
			fmt.Println("two aces! we should split!/n")
		}
	}
	return sum

}

// sort out the funds transfers and clear per-game state (hands and bets)
func settle(result int) float64 {

	net := 0.0
	switch result {
	case Win:
		net = me.bet * HousePayoffFactor
		me.funds += net
		dealer.funds -= net
	case Blackjack:
		net = me.bet * HouseBlackjackFactor
		me.funds += net
		dealer.funds -= net
	case Loss:
		net = me.bet
		dealer.funds += net
	case Push:
		me.funds += me.bet
	}

	dealer.hand = nil
	me.hand = nil
	me.bet = 0.0
	return net

}

// after player/s have stayed or busted, this is called to complete the dealer's hand
func dealer_wrap() {

	// implement the "hit 'till 17 or bust" rule if necessary
	for dealer.hand.sum() < HouseDealerStay {
		card := pop()
		dealer.hand = append(dealer.hand, card)
	}

	return

}
