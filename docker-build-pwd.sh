docker build -t devopslabs.azurecr.io/pwd/devopslabs-pwd -f Dockerfile .
docker build -t devopslabs.azurecr.io/pwd/devopslabs-pwd-l2 -f Dockerfile.l2 .
docker push devopslabs.azurecr.io/pwd/devopslabs-pwd
docker push devopslabs.azurecr.io/pwd/devopslabs-pwd-l2