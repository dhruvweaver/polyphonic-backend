FROM registry.fedoraproject.org/fedora:38

WORKDIR /app
COPY . .
RUN sudo dnf install -y golang

CMD ["go", "run", ".", "--release"]
