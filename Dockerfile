FROM golang:1.27@sha256:0ecdc2a9f6156af6451080bfe3d8382a662fcc4e209608c6f919e643453514c1
RUN mkdir /backend
COPY  --parents ["api", "cli", "internal", "go.mod", "go.sum", "app.go", "/backend/"]
RUN go build -C /backend -o gobackend

FROM node:slim
RUN mkdir /home/node/wizard
COPY --from=0 /backend/gobackend /home/node/wizard/gobackend

USER node
COPY ./command.sh /
RUN curl https://install.abra.coopcloud.tech | bash
ENV ABRA_BIN=/home/node/.local/bin/abra
# RUN $ABRA_BIN recipe fetch -a

COPY --parents web /home/node/wizard
WORKDIR /home/node/wizard/web
USER root
RUN npm install .
RUN chown -R node /home/node/
USER node

EXPOSE 5173 3000
# ENV VITE_API_URL=http://localhost:3000/api
ENV VITE_MOCK_AUTH=false
CMD [ "/command.sh" ]
