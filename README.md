# goQueue

kind of a scratch pad for loading a queue of stuff from RabbitMQ in go, hope to be able to send it instructions, and things to work on.

# ENV Variables

```
export QPASSWORD=guest
export QUSER=guest
export QHOST=127.0.0.1
```
# But My RabbitMQ is on a foreign network, locally and VPNs Suck

SSH port forward, $LOCAL = Internal address, $REMOTE = external Address.

```
ssh -L 5672:$LOCAL:5672 bsdpunk@$REMOTE -N 
ex:
ssh -L 5672:10.1.10.200:5672 bsdpunk@bsdpunksplayground.com -N 
ssh -L 5672:10.1.10.200:5672 bsdpunk@1.1.1.1 -N 
```
# TO DO

* Send Interpretable Instructions
* Set up a Go Routine Channel that Shares The Message Queue 
* Use Instructions on the Go Routine Channel
