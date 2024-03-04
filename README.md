To build and run project use
>docker-compose up --build 

To run locally replace nginx conf file content at todo_list\website\nginx\sites-enabled with:

>server {
>    listen 80;
>    server_name localhost;
>
>    return 301 https://$host$request_uri;
>}
>
>server {
>    listen 443 ssl;
>    server_name localhost;
>
>    ssl_certificate /etc/nginx/ssl/localhost.crt;
>    ssl_certificate_key /etc/nginx/ssl/localhost.key;
>
>    # Other SSL configurations
>    ssl_protocols TLSv1.2 TLSv1.3;
>    ssl_ciphers 'TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384';
>    ssl_prefer_server_ciphers off;
>    ssl_session_cache shared:SSL:10m;
>    ssl_session_timeout 10m;
>
>    location / {
>        root   /usr/share/nginx/www/todobukh.ru;
>        index  index.html index.htm;
>    }
>}
