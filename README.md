# HappyBot

## Command Stack 

The command stack is used for long running tasks that need to maintain state.

An example is a threaded conversation with a user than ends in a task being
performed.

The command stack is built up of root commands and a stack of current commands.

Commands can be either added as root commands or pushed on to the stack of
current commands. Root commands are commands that should always be available.
A module will usually have one root command that is the entry point for all its
functionality. Commands added to the stack represent discrete pieces of work
that will end at some point. *e.g. A command that PMs a question and then
saves the next number between 1 and 10 in reply via pm maybe added as part of a
command that collects OKR stats.*

When the command stack receives a message each command starting from the top of
the stack chooses whether to handle the message. When a command elects to handle
a message no futher commands are given the opportunity. If none of the current 
commands handles the message, then each root command is given a chance in the
order that the root commands were added, regardless of whether another root
command chose to handle it.


Any command may push children commands to the stack. When the child is popped
from the stack the parent command is notified. Typically a command is
responsible for popping itself from the stack. When a command is popped from
the stack all children of that command are also popped but their parents
are not notified.

With this structure complex long running command trees can be composed together
from simple primitive commands. For example the user module may add a root
command that handles any private messages from an admin user exactly equal to  
"add user", "remove user", or "update user".  

When this root command recieves an "add user" message it then creates and pushes
an add user command to the stack. This command in turn pushes a subcommand which 
when added to the stack asks the user a question and saves the input. *e.g.
"Please type the email address for the new user".* When the user responds the
command pops itself from the stack notifying the parent add user command. The 
add user command then pushes a new subcommand for a first name and so forth.
Finally, once the add user command has collected all the required information
it adds the user to the database, notifies the user, and pops itself from the
current stack.

## Users Module

- Add user
- Set user role
- Delete user
- Identify user from chat id

## OKR Module

- Be able to add OKR questions for a user
- Questions have replaceable tokens
- Answers can be boolean, range, or string
- Answers can be exported or perhaps viewed on web
- Questions are asked on a cron schedule
- Questions are posed via private message and replies parsed

