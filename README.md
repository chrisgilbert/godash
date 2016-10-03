# godash
An small amazon dash golang program to send web requests on button press.


The program itself works well, but the Dockerfile and build for Raspberry Pi is a work in progress.

# Compiling

You should be able to compile it like this:

```
git clone https://github.com/chrisgilbert/godash/
cd godash/godash
go get 
go build
```

You can then do `mv conf-example.json conf.json` and edit that file to add the appropriate details.

# Finding your dash button hardware address

I haven't provided a program to do this, but this will help (an example on Mac, it's probably eth0 on Linux):
```
tcpdump -i en0 -e | grep who-has
```
This will show some arp traffic for anything requesting an IP mac translation from the switch/router.  
You should be able to figure out what address is attached to the button fairly easily. (Mine starts with ac:)

I've only tested on Go 1.7.1.
