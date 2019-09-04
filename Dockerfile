FROM    golang:1.12

LABEL   org.label-schema.name = "Response Auth API Service"
LABEL   org.label-schema.description = "A basic email/password auth service, for use when testing."
LABEL   org.label-schema.url = "https://github.com/lavrahq/response" 
LABEL   org.label-schema.vcs-url = "https://github.com/lavrahq/response-auth"
LABEL   org.label-schema.vendor = "Lavra"
LABEL   org.label-schema.schema-version = "1.0"
LABEL   io.lavra.stack.supported = "true"

WORKDIR /app
COPY    . .

RUN     go get -u
RUN     go build -o app

ENV     PORT=8090

CMD     [ "./app" ]