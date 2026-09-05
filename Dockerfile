FROM golang:1.25@sha256:699337d620559a59b4a2bb298ad59611e535d2ee755a34cf2d2a98f37578dc80
RUN mkdir /backend
COPY  --parents ["api", "cli", "internal", "go.mod", "go.sum", "app.go", "/backend/"]
RUN go build -C /backend -o gobackend

FROM node:slim@sha256:c0753125a3789977aefe869cbebccf70e3cfd7ea84ca48547458f02e4f1d7146
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
