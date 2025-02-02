## Emoji Counter Bot for Discord 
### Using Go and docker-compose


[_Dockerfile_](backend/Dockerfile)
Dockerfile contains an example of minifying the resulting container in Go by utilizing multi-step builder.

## Deploy with docker compose (--build flag recreates if something is updated)

```shell
$ docker compose up -d --build
[+] Running 20/13
 ✔ redis Pulled                
 ✔ db Pulled 
...
=> naming to docker.io/library/discordemojicountergobot-backend
[+] Running 4/4
 ✔ Network discordemojicountergobot_default      Created
 ✔ Container discordemojicountergobot-db-1       Healthy 
 ✔ Container discordemojicountergobot-redis-1    Healthy   
 ✔ Container discordemojicountergobot-backend-1  Started  
```

## Expected result

Listing containers must show three containers running and the port mapping as below:
```shell
$ docker compose ps
NAME                                 IMAGE                              COMMAND                  SERVICE   CREATED          STATUS                 PORTS
discordemojicountergobot-backend-1   discordemojicountergobot-backend   "./bin"                  backend   19 minutes ago   Up 19 minutes
discordemojicountergobot-db-1        postgres:16-alpine                 "docker-entrypoint.s…"   db        2 hours ago      Up 2 hours (healthy)   0.0.0.0:5433->5432/tcp
discordemojicountergobot-redis-1     redis:7.4.2-alpine                 "docker-entrypoint.s…"   redis     2 hours ago      Up 2 hours (healthy)   0.0.0.0:6379->6379/tcp
```

Stop and remove the containers(and images)
```shell
$ docker-compose down --rmi all 
```