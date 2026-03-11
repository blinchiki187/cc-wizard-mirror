FROM golang:1.25
RUN mkdir /backend
COPY . /backend/
RUN go build -C /backend -o app
EXPOSE 3000

FROM node
RUN mkdir /frontend
WORKDIR /frontend
RUN git clone https://git.coopcloud.tech/BornDeleuze/coop-cloud-front .
RUN npm install
EXPOSE 5173

RUN curl https://install.abra.coopcloud.tech | bash
ENV ABRA_BIN=/root/.local/bin/abra
RUN $ABRA_BIN recipe sync


COPY ./start.sh /
COPY --from=0 /backend/app /backend/app
ENTRYPOINT [ "/start.sh" ]
