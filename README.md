# DISYS-Handin-5

Client-Server struktur (gRPC)

Passive Replication
- Single leader


Have flere servere som back-up.
Skal kunne håndtere mindst et crash af en node
- skal vi kunne håndtere crash af leader server?
- hvis ja, brug election algoritme - vælge den med opdateret data og højest id


Selve systemet:
- Bid og Result metode
- Auction skal slutte på et tidspunkt og returnere højeste Bid
- Brug Lamport timestamp (Sequential Consistency) til at bestemme orden af beskeder (Bids, Result)
- Alle replicas skal opdateres
- Sende 'heartbeat' ud fra leader til replicas, så de kan følge med om leader er død eller ikke
  - Vi tager ikke højde for hvis en client har sendt en besked lige før replicas har opdaget at leaderen er død

Ting som skal håndteres:
- To beskeder bliver sendt samtidig
- Vi får ikke noget 'ack' tilbage efter at have sendt besked
- DONE - En node fejler -> så stopper vi med at sende opdateringer til den -
- Lederen fejler -> election. evt skal clients tjekke dette


Ekstra
- håndtere en ny server kommer efter der er lavet bids
