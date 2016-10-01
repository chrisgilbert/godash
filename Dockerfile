FROM hypriot/rpi-golang

RUN mkdir ./godash
COPY godash ./godash
RUN cd ./godash && go get -u github.com/google/gopacket
RUN cd ./godash && go build
