FROM centos:7

ADD bin/apiwg /

RUN chmod 777 /apiwg

ENV PARAMS=""

ENTRYPOINT ["sh","-c","/apiwg $PARAMS"]