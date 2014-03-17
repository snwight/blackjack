blackjack
=========

Go implementation of the game of Blackjack
------

Brutally simplistic curl interface:
-

report on your available funds:

	curl http://localhost:9999/funds


increase your available funds:

	curl http://localhost:9999/deposit?amount=200


increment your current bet:

	curl http://localhost:9999/bet?amount=100


inspect your current bet:

	curl http://localhost:9999/bet


request initial pair to be dealt - dealer shows one card:

	curl http://localhost:9999/deal


request a card:
	
	curl http://localhost:9999/hit


trigger dealer play (holds at 17) and complete game:
	
	curl http://localhost:9999/stay


So-called admin api:
-

for debugging only, dump current deck contents:
	
	curl --data auth=titanoboa http://localhost:9999/show_deck


resize deck from default two:

	curl --data count=5 http://localhost:9999/size_deck


