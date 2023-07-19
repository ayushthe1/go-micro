### The Broker Service processes each request sent by the front-end microservice and sends a response back to it.
It receives request from the client and sends to the logger service.

---

#### Based on the function calls that sends to the logger-service , i pick to use any of these function that does the following  
- The broker can send requests using Api's , it sends it to the logger service and  saves it on the logger service database then displays it in the frontend 

- The broker can send events to the Queue and  send it to the logger service  the logger service database, then displays it in the frontend 

- The broker can send requets using RPC , it sends it to the logger service and saves it into logger-Service database then displays resut in the frontend 

---


#### Libraries Used 
- Uses  [chi router](https://github.com/go-chi/chi/v5) for routing
- Uses [chi](github.com/go-chi/cors) as its CORS
