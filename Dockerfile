FROM golang:1.25@sha256:699337d620559a59b4a2bb298ad59611e535d2ee755a34cf2d2a98f37578dc80
RUN mkdir /backend
COPY  --parents ["api", "cli", "internal", "go.mod", "go.sum", "app.go", "/backend/"]
RUN go build -C /backend -o gobackend

FROM node:slim@sha256:0d1e33586362bb9e608c4e751e7a9e7edb55c9ac11461ed5666af696cc461dd4
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
