# turbo-octo-potato

build:
docker build -t server-ssh .
run:
docker run -d \
  -p 2222:2222 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --name server-ssh \
  server-ssh
connect:
ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -p 2222 localhost
- -o UserKnownHostsFile=/dev/null option tells SSH to discard the server identity
- -o StrictHostKeyChecking=no option tells SSH 'yes' I want to connect immediately
