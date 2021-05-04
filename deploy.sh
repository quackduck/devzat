echo Started
scp -o StrictHostKeyChecking=no -P 4242 ./devchat.go go.sum go.mod ubuntu@150.136.142.44:~/devchat
echo Copied files
ssh -o StrictHostKeyChecking=no -p 4242 ubuntu@150.136.142.44 <<EOL # Unquote so lines are expanded
	cd ~/devchat
	go build && echo Built
	echo $SERVER_PASS | sudo -S pkill devchat && echo Killed
	sleep 2
	echo $SERVER_PASS | sudo -S pkill -9 devchat && echo Killed with SIGKILL
	echo $SERVER_PASS | nohup sudo -S HOME=/home/ubuntu ./devchat > /dev/null 2>&1 </dev/null &
	echo Started server
	disown
	exit
EOL
echo Finished
