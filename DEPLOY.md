# 服务器部署指南(Docker 一键部署)

单容器形态:一个 Go 进程同时提供 REST API(`/api/v1`)与前端页面,数据库为内嵌 SQLite(纯 Go 驱动),无外部依赖。

## 部署三步(拉取已发布镜像,推荐)

CI 推送到 main 后会自动构建多架构镜像并发布到 `ghcr.io/hodgewen/software-exam:latest`。

```bash
# 1. 装 Docker(已装跳过)
curl -fsSL https://get.docker.com | sh

# 2. 只需拿到 docker-compose.yml(整个仓库 clone 下来也行)
git clone https://github.com/HodgeWen/software-exam.git && cd software-exam

# 3. 拉镜像并启动
docker compose pull && docker compose up -d
```

浏览器访问 `http://<服务器IP>:8080`,注册一个账号即可开始刷题。首次启动会自动建表并导入全部内置题库。

**镜像拉取权限**:GHCR 包默认私有。若 `docker compose pull` 报无权限,二选一:
- 在 GitHub 仓库页 → Packages → software-exam → Package settings → Change visibility 改为 Public(公开仓库的镜像,改一次即可);
- 或保持私有,服务器上 `docker login ghcr.io`(用户名 + 有 read:packages 权限的 PAT)。

## 备选:服务器自构建

不需要发布镜像时,直接在服务器上构建(效果相同):

```bash
docker compose up -d --build
```

## 常用配置

在仓库根目录放一个 `.env` 文件(与 docker-compose.yml 同目录):

```env
HOST_PORT=8080                    # 对外端口
EXAM_JWT_SECRET=一段随机长字符串   # JWT 密钥,固定后重启不丢登录态
```

改完执行 `docker compose up -d` 生效。密钥建议 `openssl rand -hex 32` 生成。

## 日常运维

```bash
docker compose logs -f            # 看日志
docker compose restart            # 重启
docker compose down               # 停止(数据保留)
```

- **数据备份**:全部数据在 `./data/exam.db` 一个文件,停服后直接拷贝该目录即可
- **升级版本**:`git pull && docker compose pull && docker compose up -d`(自构建部署则用 `up -d --build`),题库种子幂等导入,新题库自动补入
- **数据迁移**:把整个 `data/` 目录拷到新服务器同位置再启动

## 说明

- 反向代理:如需 HTTPS/域名,前面挂 Nginx/Caddy 转发到 `127.0.0.1:8080` 即可,无需改应用
- 开发期形态不变:前端 `vp dev` 走 Vite 代理,后端不设 `EXAM_STATIC_DIR` 时只提供 API
