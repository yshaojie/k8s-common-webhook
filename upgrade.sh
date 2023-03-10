tag=$(date '+%m%d%H%M%S')
make docker-build IMG=common-webhook:${tag}
kind load docker-image common-webhook:${tag}
sleep 3
make deploy IMG=common-webhook:${tag}