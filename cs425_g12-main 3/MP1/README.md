# CS425_MP1_G12

# Running Server:
For every machine, run the following commands:
1. cd server
2. go run server.go 'machine no.'
    Machine no: 1 (for searching on the provided log files) or 01 (for searching on generated log files)

# Running Client:
For the client machine, run the following commands:
1. cd client
2. go run client.go 'mode'
    Mode: demo (for running on the provided log files) or unit_test (for running on generated log files)
3. enter the grep command after servers are connected

# Running Unit Tests:
## Generate logs first
1. cd unit_tests
2. python3 log_gen_tests.py 'username'

## Run unit tests
1. cd unit_tests
2. python3 pattern_checking.py