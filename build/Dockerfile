FROM java:8-jdk

ARG USER=cerana
ARG GROUP=cerana
ARG UID=10000
ARG GID=10000

ARG JENKINS_VERSION=2.4
ARG JENKINS_SHA=c33df09d0310fc2f62c9d14b84d06fe8529307ea
ARG JENKINS_WEB_PORT=8080
ARG JENKINS_SLAVE_PORT=50000
ARG JENKINS_HOME=/home/cerana/.jenkins

USER root
RUN apt-get update && \
  apt-get install -y git curl zip bzip2 adduser graphviz sudo time nano man

#+
# Normally Jenkins runs as the jenkins user. In this case it defaults to
# cerana.
#-
RUN adduser --disabled-password --gecos '' --uid ${UID} ${USER}

# Use tini as subreaper in Docker container to adopt zombie processes
ENV TINI_SHA 066ad710107dc7ee05d3aa6e4974f01dc98f3888

RUN curl -fsSL https://github.com/krallin/tini/releases/download/v0.5.0/tini-static \
  -o /bin/tini && chmod +x /bin/tini \
  && echo "$TINI_SHA  /bin/tini" | sha1sum -c -

#+++++++++++++++
# For Jenkins
# NOTE: Much of this is derived from: https://github.com/jenkinsci/docker
#+++++++++++++++

#+
# Jenkins home directory is a volume, so configuration and build history
# can be persisted and survive image upgrades
#-
VOLUME ${JENKINS_HOME}

#+
# Plugins and config for this are here.
#-
COPY jenkins.sh /usr/local/bin/jenkins.sh
COPY plugins.sh /usr/local/bin/plugins.sh
RUN mkdir -p /usr/share/jenkins/ref/init.groovy.d
COPY init.groovy /usr/share/jenkins/ref/init.groovy.d/tcp-slave-agent-port.groovy

RUN curl -fsSL \
  http://repo.jenkins-ci.org/public/org/jenkins-ci/main/jenkins-war/${JENKINS_VERSION}/jenkins-war-${JENKINS_VERSION}.war \
  -o /usr/share/jenkins/jenkins.war && \
  echo "$JENKINS_SHA  /usr/share/jenkins/jenkins.war" | sha1sum -c -

RUN chown -R ${USER} ${JENKINS_HOME} /usr/share/jenkins/ref

EXPOSE ${JENKINS_WEB_PORT}
EXPOSE ${JENKINS_SLAVE_PORT}

ENV JENKINS_UC https://updates.jenkins.io
ENV JENKINS_VERSION ${JENKINS_VERSION}
ENV JENKINS_SHA ${JENKINS_SHA}
ENV COPY_REFERENCE_FILE_LOG ${JENKINS_HOME}/copy_reference_file.log

#+++++++++++++++
# For builds using Nix.
#+++++++++++++++
RUN mkdir -m 777 /nix
COPY nix-builder.sh /home/${USER}/

USER ${USER}
ENV LC_ALL C
ENV USER ${USER}
WORKDIR /home/${USER}
RUN curl https://nixos.org/nix/install | sh

#+
# For container startup.
#-
ENTRYPOINT ["/bin/tini", "--", "/usr/local/bin/jenkins.sh"]
