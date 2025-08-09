FROM gradle:8.7.0-jdk17

RUN apt-get update && \
    apt-get install -y make build-essential && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . /app

RUN gradle --no-daemon build

CMD ["make", "run", "CLASS=BasicStreams"]
