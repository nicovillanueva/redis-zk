FROM alpine

ADD requirements.txt /etc/watcher/
RUN apk --update add python3 && \
    pip3 install -r /etc/watcher/requirements.txt
ADD watcher.py /etc/watcher

ENTRYPOINT ["python3","/etc/watcher/watcher.py"]
