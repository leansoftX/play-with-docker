docker build -t devopslabs.azurecr.io/pwd/devopslabs-pwk -f Dockerfile .
docker build -t devopslabs.azurecr.io/pwd/devopslabs-pwk-l2 -f Dockerfile.l2 .
docker push devopslabs.azurecr.io/pwd/devopslabs-pwk
docker push devopslabs.azurecr.io/pwd/devopslabs-pwk-l2