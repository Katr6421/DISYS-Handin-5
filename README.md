# DISYS-Handin-5

## How to run

First open three seperate terminals from the server directory

In terminal 1 write:

   ```sh
    go run . 0  
  ```

In terminal 2 write:

  ```sh
   go run . 1 
  ```

In terminal 3 write:

  ```sh
    go run . 2  
  ```

To start a client, open a new terminal from the client directory and write:

  ```sh
    go run .  
  ```

You can open as many clients as you wish

## Auction

* The auction is open for 5 minutes from the first server is started.
* Clients can make bids by writing a number in their terminal
* Clients can request the status of the aution by writing "status" in the terminal

## Log

The operations in the program will be logged in the file called `log.log`

## Notes

Client-Server struktur (gRPC) men hvor serverne bruger peer-to-peer

Generelt:

* Auction slutter 5 minutter efter leader replica er lavet
* Der sendes heartbeat hvert 15. sekund

Passive Replication

* Single leader
* Sender heartbeat hvert 15 sekunder og smider de backups ud, der ikke svarer
* Opdaterer backups hver gang der sker ændringer i auktionen

Crash af backup replica

* En backup replica fjernes hvis den:

* Ikke svarer på heartbeat

* Ikke svarer på update

MANGLER:

* Crash af leader. (enten timeout på at modtage heartbeat eller at det opdages når client prøver at sende en besked, og så vælges der en ny leder?)
* To beskeder bliver sendt samtidig (lamport?), Sequential Consistency (er det fixet med lock??)

HUSK I RAPPORTEN

* Client vs frontend
* Client ville normalt kunne spørge om status fra any backup replica, men her kan de kun fra lederen (kan kun connectes til én server)
