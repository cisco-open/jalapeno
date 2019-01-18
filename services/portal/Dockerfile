FROM node:alpine as assets
WORKDIR /data/
RUN npm install @creative-tim-official/argon-dashboard-free

FROM python:3-alpine
WORKDIR /data/
COPY src/requirements.txt .
RUN apk add --no-cache --virtual .build-deps gcc musl-dev
RUN pip3 install --no-cache-dir -r requirements.txt
RUN apk del .build-deps gcc musl-dev
COPY src/ .
COPY --from=assets /data/node_modules/@creative-tim-official/argon-dashboard-free/assets static/argon
EXPOSE 80
ENTRYPOINT [ "python3" ]
CMD [ "app.py" ]
