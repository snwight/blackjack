package main

import (
	"fmt"
	"math/rand"
	"time"
	"github.com/hoisie/web"
)


/*
    data type declarations
 */
type Card struct {
    value int
    name string
}

type Hand []Card
type Deck []Card
type Player struct {
	name string
	hand Hand
}


/*
    "global" data objects
 */

// initialized sorted deck
var fullDeck = Deck {
	{1, "CA"}, {2, "C2"}, {3, "C3"}, {4, "C4"}, {5, "C5"}, {6, "C6"}, {7, "C7"}, 
	{8, "C8"}, {9, "C9"}, {10, "C10"}, {10, "CJ"}, {10, "CQ"}, {10, "CK"},
	{1, "DA"}, {2, "D2"}, {3, "D3"}, {4, "D4"}, {5, "D5"}, {6, "D6"}, {7, "D7"}, 
	{8, "D8"}, {9, "D9"}, {10, "D10"}, {10, "DJ"}, {10, "DQ"}, {10, "DK"},
	{1, "HA"}, {2, "H2"}, {3, "H3"}, {4, "H4"}, {5, "H5"}, {6, "H6"}, {7, "H7"}, 
	{8, "H8"}, {9, "H9"}, {10, "H10"}, {10, "HJ"}, {10, "HQ"}, {10, "HK"},
	{1, "SA"}, {2, "S2"}, {3, "S3"}, {4, "S4"}, {5, "S5"}, {6, "S6"}, {7, "S7"}, 
	{8, "S8"}, {9, "S9"}, {10, "S10"}, {10, "SJ"}, {10, "SQ"}, {10, "SK"},
}
var deck = fullDeck

var dealer = Player{"dealer", nil}
var me = Player{"me", nil}


/*
    load the callbacks and start the webgo server
 */
func main() {
	rand.Seed(time.Now().UnixNano())
	web.Get("/deal", deal)
	web.Get("/hit", hit)
	web.Post("/bet", bet)
	web.Run("0.0.0.0:9999")
}


/*
    API callbacks
 */

// curl http://localhost:9999/deal
func deal() Hand {
	// for now reload and shuffle deck for every hand!
	shuffle()
	dealer.hand = nil
	me.hand = nil
	dealer.hand = append(dealer.hand, deck[0], deck[2])
	me.hand = append(me.hand, deck[1], deck[3])
 	deck = deck[4:]
	fmt.Printf("dealer showing: %+v, you have: %+v\n", dealer.hand, me.hand)
	return me.hand
}

// curl http://localhost:9999/hit
func hit() Hand {
	card := deck[0]
	deck = deck[1:]
	me.hand = append(me.hand, card)
	fmt.Printf("hit: %+v\n", card)
	return me.hand
}

// curl --data "amount=XXX" http://localhost:9999/bet
func bet(ctx *web.Context) string { 
	return "bet: " + ctx.Params["amount"] + "\n"
}


/*
    utility functions
 */

// single deck game is default - append to slice as needed
func shuffle() Deck {
	deck = fullDeck
	l := len(deck)
	// Algorithm P, Knuth/Durstenfeld
	for i := l - 1; i > 0; i-- {
		j := rand.Intn(l - i)
		tmp := deck[j]
		deck[j] = deck[i]
		deck[i] = tmp
	}
	fmt.Printf("shuffled deck: %+v\n", deck)
	return deck
}

