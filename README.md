# Lightblocks

Project implements client and server with ordered map implementation according to the assignment.

## Prerequisites

Before running the application or tests, ensure that you have Go installed on your system. You can download and install it from [https://golang.org/dl/](https://golang.org/dl/).

## Running the Application

After cloning the application and downloading necessary packages we can run server and client as follows (example command list file is provider). Command execution (the output) will be writtein in `output.txt` file in the working directory. Please note that outut
file might be misleading to understand if the application works correctly as we are writing in the file from multiple go routines.
the only guarantee is that the command `getAllItems` will write items in the order they were inserted but it could be corrupted/interrupted by other `getAllItems` or `getItem` command and the list might not be the same as expected.

   ```bash
   git clone git@github.com:shotasilagadze/lightblocks.git
   cd lightblocks
   go run server/main.go amqp://guest:guest@localhost:5672/ output.txt


   go run client/main.go -file commands_1.json
   go run client/main.go -file commands_2.json
   go run client/main.go -file commands_3.json
   .
   .
   .
   go run client/main.go -file commands_n.json
   ```
## Running tests
    
    go test ./...
