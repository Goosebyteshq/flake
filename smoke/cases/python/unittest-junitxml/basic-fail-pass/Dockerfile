FROM python:3.12-alpine
WORKDIR /app
COPY . .
RUN pip install --no-cache-dir unittest-xml-reporting
CMD ["sh", "-c", "python -m xmlrunner discover -s . -p \"test_*.py\" -o reports >/tmp/test.log 2>&1 || true; cat reports/*.xml"]
