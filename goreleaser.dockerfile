FROM scratch
COPY op-plasma-celestia /usr/bin/op-plasma-celestia
ENTRYPOINT ["/usr/bin/op-plasma-celestia"]
