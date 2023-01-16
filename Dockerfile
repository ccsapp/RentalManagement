FROM alpine

ARG projectname=main
ENV projectname="${projectname}"

COPY bin/* /usr/app/

EXPOSE 80

ENTRYPOINT ["sh", "-c", "/usr/app/$(ls /usr/app)"]
