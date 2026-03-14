# coop-cloud-backend
A Go service that exposes RESTful API endpoints to manage Abra programmatically.  
Integrates with https://git.coopcloud.tech/BornDeleuze/coop-cloud-front.  

## Starting the service with Docker
TODO

## Getting started

- Edit the front-end application to turn off mock mode in `src/routes/Authenticated/Apps/App.tsx` and `src/routes/Authenticated/Apps/Apps.tsx` and `src/routes/Authenticated/Dashboard/Dashboard.tsx`
- Launch the front-end application `npm run dev`  
- Start this Go app `go run .`  
- Navigate to the React App (http://localhost:5173)  

> [!WARNING]
> This is an extremely early prototype, only viewing/deploying apps is supported
> and may fail for your local machine