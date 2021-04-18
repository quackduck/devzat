echo Started
mkdir ~/.ssh
echo $PRIVKEY > ~/.ssh/id_rsa
echo $PUBKEY > ~/.ssh/id_rsa.pub
echo Wrote private key
scp -P 4242 ./devchat.go go.sum go.mod ishan@34.75.6.116:~/devchat
echo Copied files
ssh -T -P 4242 -l ishan 34.75.6.116 <<'EOL'
	cd ~/devchat
	go build
	echo Built
	echo $SERVER_PASS | sudo -S pkill devchat
	echo Killed
	echo $SERVER_PASS | sudo -S HOME=/home/ishan ./devchat &; disown
	echo Started server
EOL
echo Finished
