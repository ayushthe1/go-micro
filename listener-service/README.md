##  Listener Service

### This Service handles the message queues received from the broker service .

- The listener service is a message broker for receiving and handling messages(request) from the broker service and calls the right service to hadle that particular request .

The User makes a request to the broker  & pushes the request to the rabbitmq, the listener-service gets a request from the queue to either , authenticate, send an email, and log something, if a user tries to login it sends a request to the authentication service and takes action based on the result attempt.
