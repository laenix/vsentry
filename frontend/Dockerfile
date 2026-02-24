# Production - use pre-built dist
FROM nginx:alpine

# Copy pre-built assets
COPY dist /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
