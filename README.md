blackjack
=========

Go implementation of the game of Blackjack
------

Brutally simplistic curl interface:
-

create player/s (persistent for life of server instance only) and optionally deposit magic funds into their account:

	curl http://localhost:9999/create_player?name=NAME&amount=XXX


any or all players can now each repeatedly create and play personal games simultaneously using player NAME as key:
	
	curl http://localhost:9999/create_game?name=NAME


report on your available funds:

	curl http://localhost:9999/funds?name=NAME


increase your available funds:

	curl http://localhost:9999/deposit?name=NAME&amount=200


increment your current bet:

	curl http://localhost:9999/bet?name=NAME&amount=100


inspect your current bet:

	curl http://localhost:9999/bet?name=NAME


request initial pair to be dealt - dealer shows one card:

	curl http://localhost:9999/deal?name=NAME


request a card:
	
	curl http://localhost:9999/hit?name=NAME


trigger dealer play (holds at 17) and complete game:
	
	curl http://localhost:9999/stay?name=NAME


So-called admin api:
-

for debugging only, dump current deck contents:
	
	curl --data auth=titanoboa http://localhost:9999/show_deck?name=NAME


resize deck from default two:

	curl --data count=5 http://localhost:9999/size_deck?name=NAME


