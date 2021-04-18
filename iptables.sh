#!/bin/bash
IPT=/sbin/iptables
# Max connection in seconds
TIME_PERIOD=100
# Max connections per IP
BLOCKCOUNT=100

# default action can be DROP or REJECT
DACTION="DROP"
$IPT -A INPUT -p tcp --dport 80 -i eth0 -m state --state NEW -m recent --set
$IPT -A INPUT -p tcp --dport 80 -i eth0 -m state --state NEW -m recent --update --seconds $TIME_PERIOD --hitcount $BLOCKCOUNT -j $DACTION