FROM redis:3.2.1-alpine

ADD requirements.txt /etc/redis/
RUN apk --update add net-tools python3 && \
    pip3 install -r /etc/redis/requirements.txt
ADD sentinel-example.conf /etc/redis/
ADD sentinel.conf /etc/redis/
ADD wrapper.py /etc/redis/

ENTRYPOINT ["python3", "/etc/redis/wrapper.py"]
