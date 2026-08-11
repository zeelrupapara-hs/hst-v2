# Build Stage
FROM node:20-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm install --force
COPY . .
# baked at build time: where the API and the socket live for this environment
ARG VITE_BASE_URL=""
ARG VITE_SOCKET_URL=""
ENV VITE_BASE_URL=$VITE_BASE_URL VITE_SOCKET_URL=$VITE_SOCKET_URL
RUN npm run build
 
# Production Stage
FROM nginx:stable-alpine AS production
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]