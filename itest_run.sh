docker-compose -f itest-docker-compose.yml down --remove-orphans --volumes
sudo rm -rf /Users/Nursultan/volumes/test/czips
docker-compose -f itest-docker-compose.yml up --abort-on-container-exit --build
docker-compose -f itest-docker-compose.yml down --remove-orphans --volumes
