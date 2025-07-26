# Deploying OCLI SSH to Railway

This guide shows how to deploy the OCLI SSH server to Railway for public access.

## Quick Railway Deployment

### 1. Prerequisites

- GitHub account
- Railway account (sign up at [railway.app](https://railway.app))
- Your ocli repository on GitHub

### 2. Deploy to Railway

#### Step 1: Create Project
1. Go to [railway.app](https://railway.app) and sign in
2. Click "New Project"
3. Choose "Deploy from GitHub repo"
4. Connect your GitHub account if needed
5. Select your ocli repository

#### Step 2: Configure Service
1. Railway will detect the Dockerfile automatically
2. Wait for the initial build to complete (may fail - that's okay)

#### Step 3: Add Persistent Volume
1. In your Railway project dashboard, click on your service
2. Go to the "Variables" tab
3. Scroll down to the "Volumes" section
4. Click "New Volume"
5. Configure:
   - **Mount Path**: `/data`
   - **Size**: 1GB (can increase later)
6. Click "Add Volume"

#### Step 4: Set Environment Variables
In the "Variables" tab, add these environment variables:
```
OCLI_SSH_AUTO_REGISTER=true
OCLI_SSH_DATA_DIR=/data
```

#### Step 5: Deploy
1. Go to "Deployments" tab
2. Trigger a new deployment (or push a commit to trigger auto-deploy)
3. Wait for build to complete
4. Your service will be available at `your-project.railway.app`

### 3. Connect to Your Server

Once deployed, users can connect with:

```bash
ssh alice@your-project.railway.app -p 443
```

**Note**: Railway maps the internal port to 443 (HTTPS port) for external access.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 2222 | Port to bind (Railway sets this automatically) |
| `OCLI_SSH_AUTO_REGISTER` | false | Enable auto-registration of new users |
| `OCLI_SSH_DATA_DIR` | /var/lib/ocli-ssh | Directory for user data |
| `OCLI_SSH_HOST` | 0.0.0.0 | Host to bind to |

## Auto-Registration Flow

With `OCLI_SSH_AUTO_REGISTER=true`:

1. User runs: `ssh alice@your-domain.com`
2. Server extracts SSH public key from connection
3. If user doesn't exist, creates new user with that key
4. User gets access to personal OCLI instance

## Custom Domain Setup

1. **In Railway Dashboard**:
   - Go to your project
   - Click "Settings" → "Domains"
   - Add your custom domain (e.g., `ocli.yourdomain.com`)

2. **DNS Configuration**:
   ```
   CNAME: ocli.yourdomain.com → your-project.railway.app
   ```

3. **Connect**:
   ```bash
   ssh alice@ocli.yourdomain.com
   ```

## Volume Persistence

Railway provides persistent volumes for data storage:

1. **Create Volume**:
   - Go to your service in Railway dashboard
   - Click "Variables" tab
   - Scroll to "Volumes" section
   - Click "New Volume"

2. **Configure Volume**:
   - **Mount Path**: `/data`
   - **Size**: Start with 1GB, can be increased later
   - Click "Add Volume"

3. **Verify Configuration**:
   - Ensure `OCLI_SSH_DATA_DIR=/data` environment variable is set
   - User data will persist across deployments and restarts

**Important**: Without a volume, user data will be lost on each deployment.

## Monitoring

### View Logs
```bash
# In Railway dashboard
railway logs
```

### Check Server Status
```bash
# Connect and check
ssh admin@your-domain.com
```

## Security Considerations

1. **Host Key Persistence**: The Dockerfile generates a host key on first run. Consider using Railway's persistent volumes to maintain the same host key.

2. **Rate Limiting**: Railway provides DDoS protection, but consider implementing connection rate limiting for production use.

3. **User Management**: With auto-registration, monitor user creation and disk usage.

## Troubleshooting

### Connection Issues

1. **Check Railway Logs**:
   - View deployment logs in Railway dashboard
   - Look for startup errors

2. **Test Local Connection**:
   ```bash
   ssh -v alice@your-domain.com
   ```

3. **Common Issues**:
   - Port 22 vs 443: Railway routes external traffic to port 443
   - SSH key format: Ensure your SSH key is in the correct format
   - Username: Use any username - it will be auto-created

### Build Issues

1. **Dockerfile Location**: Ensure the Dockerfile is in `cmd/ocli-ssh/`
2. **Go Modules**: Check that go.mod is correctly configured
3. **Dependencies**: Verify all imports are available

## Cost Estimation

Railway pricing:
- **Hobby Plan**: $5/month for small projects
- **Pro Plan**: $20/month + usage for production

The SSH server is lightweight and should run comfortably on the hobby plan for moderate usage.

## Alternative Deployment Options

### 1. Docker + VPS

```bash
# Build and run on any VPS
docker build -t ocli-ssh cmd/ocli-ssh/
docker run -d -p 2222:2222 -v ocli-data:/data -e OCLI_SSH_AUTO_REGISTER=true ocli-ssh
```

### 2. Heroku

Create `heroku.yml`:
```yaml
build:
  docker:
    web: cmd/ocli-ssh/Dockerfile
```

### 3. AWS/GCP/Azure

Use their container services with the provided Dockerfile.

## Support

For issues:
1. Check the logs in Railway dashboard
2. Test locally with Docker
3. Open an issue in the GitHub repository