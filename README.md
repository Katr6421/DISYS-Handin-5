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