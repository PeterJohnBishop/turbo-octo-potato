# turbo-octo-potato

build:
docker build -t server-ssh .
run:
docker run -d -p 2222:2222 --name server-ssh server-ssh
connect:
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -p 2222 localhost