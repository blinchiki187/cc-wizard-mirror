FROM golang:1.25
RUN mkdir /backend
COPY . /backend/
RUN go build -C /backend -o app
EXPOSE 3000

FROM node
RUN mkdir /frontend
WORKDIR /frontend
RUN git clone https://git.coopcloud.tech/BornDeleuze/coop-cloud-front
EXPOSE 5173

COPY ./start.sh /
COPY --from=0 /backend/app /backend/app
ENTRYPOINT [ "/start.sh" ]
