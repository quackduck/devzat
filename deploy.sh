echo $PRIVKEY > ~/.ssh/id_rsa
scp -P 4242 ./devchat.go go.sum go.mod ishan@34.75.6.116:~/devchat
ssh -T -P 4242 ishan@34.75.6.116 <<'EOL'
	cd ~/devchat
	go build
	echo Built
	echo $SERVER_PASS | sudo -S pkill devchat
	echo Killed
	echo $SERVER_PASS | sudo -S HOME=/home/ishan ./devchat &; disown
	echo Started server
EOL
