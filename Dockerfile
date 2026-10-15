FROM golang:1.27@sha256:820153096fda4de6ac1b7194d12e416b7dd8f4e9a33d17e0345b111b945162d8
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
