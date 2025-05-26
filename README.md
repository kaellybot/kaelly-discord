# kaelly-discord

[![CI](https://github.com/kaellybot/kaelly-discord/actions/workflows/ci.yml/badge.svg)](https://github.com/kaellybot/kaelly-discord/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/kaellybot/kaelly-discord/branch/main/graph/badge.svg)](https://codecov.io/gh/kaellybot/kaelly-discord) 


Application to interact with Discord written in Go 

## Local Development

You will probably need these docker images to make it work properly:

```Bash
docker run --name mysql --restart=always -p 3306:3306 -e MYSQL_ROOT_HOST=% -e MYSQL_ROOT_PASSWORD=password -d mysql/mysql-server:latest 
docker run --name phpmyadmin --restart=always -d --link mysql:db -p 9001:80 phpmyadmin/phpmyadmin:latest
docker run --name rabbitmq --restart=always -p 15672:15672 -p 5672:5672 -d rabbitmq:3-management:latest
docker run -p 6379:6379 --name redis -d redis
```