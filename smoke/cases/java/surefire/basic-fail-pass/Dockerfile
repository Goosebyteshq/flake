FROM maven:3.9.9-eclipse-temurin-17
WORKDIR /app
COPY pom.xml pom.xml
COPY src src
CMD ["sh", "-c", "mvn -B -Dstyle.color=never test 2>&1 || true"]
