# coop-cloud-backend
A Go service that exposes RESTful API endpoints to manage Abra programmatically.  
Integrates with https://git.coopcloud.tech/BornDeleuze/coop-cloud-front.  


## Getting started

- Edit the front-end application to turn off mock mode in `src/routes/Authenticated/App.tsx` and `src/routes/Authenticated/Apps.tsx`.   
- Launch the front-end application `npm run dev`  
- Start this Go app `go run .`  
- Navigate to the React App (http://localhost:5173)  

> [!WARNING]
> This is an extremely early prototype, only viewing apps is supported 
> and may fail for your local machine