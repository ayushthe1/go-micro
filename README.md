# Go-micro

This is a project consisting of 6 different microservices written in GoLang. These services communicate with each other using REST, RPC & gRPC APIs.
This project can be deployed either via a docker swarm or a minikube cluster.

### To deploy via Docker swarm 

1. Ensure you have docker installed and running on your system.

2. In your terminal ,type `sudo nano /etc/hosts` and enter your password.

3. In the file ,add a line `127.0.0.1      backend` . The updated file should look something like this : 
<br>

 ![image](https://github.com/ayushthe1/go-micro/assets/114604338/8f5c0541-a892-4b8e-90e3-fee70b14fed6)

4. Clone this repo

5. From inside the repo, cd into `/project` directory.

6. Type command `make deploy_swarm` in the terminal.

7. The project will be deployed on docker swarm.



### More details coming soon ....😙
