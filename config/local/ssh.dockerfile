FROM sickp/centos-sshd

COPY ./config/local/.ssh /root/.ssh
RUN chmod -R og-wrx /root/.ssh