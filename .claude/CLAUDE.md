# My Auth

A private second-factor-authenticator.

## Function
This is a CLI application what will give user the authentication code.
Available switches:
- --list (-l): to list the accounts
- --account (-a): show the authentication code for the account
- --import (-i): import from the tokens in the given .csv file
- --copy (-c): an extra switch next to --account, when given then the code will also be copied to clipboard 
- --new (-n): add new account

When importing tokens, ask user for the password. Once imported, user does not have to enter the password
anymore when asking for the list or when asking for the code itself.

## Requirements
- When accounts are imported, save the data encripted
- Should work (compiled for) Windows, MacOs (arm + intel), Linux

