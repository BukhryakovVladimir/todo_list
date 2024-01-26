# Use the official Nginx base image
FROM nginx:latest

# Remove the default Nginx configuration file
RUN rm /etc/nginx/conf.d/default.conf

# Copy the custom Nginx configuration
COPY ./website/nginx/sites-enabled/todobukh.conf /etc/nginx/conf.d/

# Create a directory for the frontend code
RUN mkdir -p /usr/share/nginx/www/todobukh.ru

# Copy the frontend code into the container
COPY ./website/todobukh.ru /usr/share/nginx/www/todobukh.ru

# Set permissions if necessary
# RUN chown -R nginx:nginx /usr/share/nginx/www/todobukh.ru

# Expose ports
EXPOSE 80 443

# CMD and ENTRYPOINT statements if needed
# CMD ["nginx", "-g", "daemon off;"]
