# coop-cloud-backend
A Go service that exposes RESTful API endpoints to manage Abra programmatically.  
Integrates with https://git.coopcloud.tech/BornDeleuze/coop-cloud-front.  

## Starting the service with Docker
TODO

## Getting started

- Edit the front-end application to turn off mock mode in `src/routes/Authenticated/Apps/App.tsx` and `src/routes/Authenticated/Apps/Apps.tsx` and `src/routes/Authenticated/Dashboard/Dashboard.tsx` and `src/routes/Authenticated/Recipes/Recipes.tsx`
- Launch the front-end application `npm run dev`  
- Start this Go app `go run .`  
- Navigate to the React App (http://localhost:5173)  

> [!WARNING]
> This is an extremely early prototype, only viewing/deploying apps is supported
> and may fail for your local machine

## Notes:
`api/` is a deprecated path, currently has no function

# Roadmap
## General Roadmap:
Currently refactored the backend to be much more Abra CLI reliant. This is ok but is not ideal from an engineering standpoint (probably).  
First step is just to polish this backend & frontend, estimate 1 - 2 months to be in a fairly usable state.  
Write test suite for the full stack combination.  
Then we want to package this app into a dockerized form, easily to deploy and use on one's own server.   
Slowly introduce full Abra CLI functionality into app.  
Repackaging
## Basic Abra Features
| Feature      | Status | Follow-up |
| ----------- | ----------- | ----------- |
| More options for app deploy (esp chaos) | TODO | |
| View and pull in recipes from catalogue | In Progress | |
| Deploy should show deployment status similar to Abra CLI | In Progress | |
| App screen should display services associated with app and their status | TODO | | 
| Easily show service logs | In Progress | |
| Modify App config | TODO | |
| Manage updates and rollbacks | TODO | |
| Manage secrets and Abra commands | In Progress | |
