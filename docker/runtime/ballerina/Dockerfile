FROM bitnami/minideb

# Install ballerina
RUN install_packages curl && \
    curl -kLO https://product-dist.ballerina.io/downloads/0.970.0/ballerina-platform-linux-installer-x64-0.970.0.deb && \
    dpkg -i ballerina-platform-linux-installer-x64-0.970.0.deb && \
    rm ballerina-platform-linux-installer-x64-0.970.0.deb

USER 1000

CMD [ "ballerina", "run", "/ballerina/files/kubeless.tpl.bal" ]
