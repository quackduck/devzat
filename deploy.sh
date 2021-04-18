echo Started
#mkdir ~/.ssh
#echo $PRIVKEY > ~/.ssh/id_rsa
# sed -i -e "s#\\\\n#\n#g" ~/.ssh/id_rsa # https://titanwolf.org/Network/Articles/Article?AID=9397221a-10ca-49af-a8ca-5008c5192ae2#gsc.tab=0
#echo $PUBKEY > ~/.ssh/id_rsa.pub
#chmod 600 ~/.ssh/*
scp -o StrictHostKeyChecking=no -P 4242 ./devchat.go go.sum go.mod ishan@34.75.6.116:~/devchat
echo Copied files
ssh -o StrictHostKeyChecking=no -T -p 4242 ishan@34.75.6.116 <<EOL # Unquote so lines are expanded
	cd ~/devchat
	go build && echo Built
	echo $SERVER_PASS | sudo -S pkill devchat && echo Killed
	echo $SERVER_PASS | sudo -S HOME=/home/ishan ./devchat &
	disown
	echo Started server
EOL
#rm -r ~/.ssh
echo Finished
