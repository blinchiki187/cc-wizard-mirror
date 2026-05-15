FROM golang:1.25
RUN mkdir /backend
COPY  --parents ["api", "cli", "internal", "go.mod", "go.sum", "app.go", "/backend/"]
RUN go build -C /backend -o gobackend

FROM node
RUN mkdir /home/node/wizard
COPY --from=0 /backend/gobackend /home/node/wizard/gobackend

COPY --parents web /home/node/wizard
WORKDIR /home/node/wizard/web
RUN npm install .

USER node
RUN curl https://install.abra.coopcloud.tech | bash
ENV ABRA_BIN=/home/node/.local/bin/abra
RUN $ABRA_BIN recipe fetch -a

COPY ./start.sh /
COPY dot_abra /home/node/.abra/
COPY ssh_config /home/node/.ssh/config
COPY id_ed25519_* /home/node/.ssh/

USER root
RUN chown -R node /home/node/
USER node

EXPOSE 5173 3000
# ENV VITE_API_URL=http://localhost:3000/api
CMD [ "/start.sh" ]
