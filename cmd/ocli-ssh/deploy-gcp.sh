#!/bin/bash
set -e

# GCP deployment script for OCLI SSH server
# Usage: ./deploy-gcp.sh

PROJECT_ID="gen-lang-client-0348675623"
INSTANCE_NAME="ocli-ssh-server"
ZONE="us-central1-a"
MACHINE_TYPE="e2-micro"  # Free tier eligible
DISK_SIZE="10GB"

echo "🚀 Deploying OCLI SSH server to Google Cloud..."

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo "❌ gcloud CLI not found. Please install it: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Set project
echo "📋 Setting GCP project to: $PROJECT_ID"
gcloud config set project $PROJECT_ID

# Build the binary for Linux
echo "🔨 Building Linux binary..."
GOOS=linux GOARCH=amd64 go build -o ocli-ssh-server .

# Create startup script
cat > startup-script.sh << 'EOF'
#!/bin/bash
set -e

# Update system
apt-get update
apt-get install -y ca-certificates

# Create ocli user
useradd -m -s /bin/bash ocli || true

# Create data directory
mkdir -p /opt/ocli-ssh/data
chown -R ocli:ocli /opt/ocli-ssh

# Create systemd service
cat > /etc/systemd/system/ocli-ssh.service << 'SYSTEMD_EOF'
[Unit]
Description=OCLI SSH Server
After=network.target

[Service]
Type=simple
User=ocli
WorkingDirectory=/opt/ocli-ssh
ExecStart=/opt/ocli-ssh/ocli-ssh-server --host=0.0.0.0 --port=2222 --data-dir=/opt/ocli-ssh/data --auto-register
Restart=always
RestartSec=5
Environment=OCLI_SSH_AUTO_REGISTER=true

[Install]
WantedBy=multi-user.target
SYSTEMD_EOF

# Enable and start service
systemctl daemon-reload
systemctl enable ocli-ssh
systemctl start ocli-ssh

echo "✅ OCLI SSH server installed and running"
EOF

# Create the VM instance
echo "🔧 Creating Compute Engine instance..."
gcloud compute instances create $INSTANCE_NAME \
    --zone=$ZONE \
    --machine-type=$MACHINE_TYPE \
    --boot-disk-size=$DISK_SIZE \
    --boot-disk-type=pd-standard \
    --boot-disk-device-name=$INSTANCE_NAME \
    --image-family=ubuntu-2204-lts \
    --image-project=ubuntu-os-cloud \
    --metadata-from-file startup-script=startup-script.sh \
    --tags=ocli-ssh-server \
    --scopes=https://www.googleapis.com/auth/cloud-platform

# Create firewall rule for SSH port
echo "🔥 Creating firewall rule..."
gcloud compute firewall-rules create allow-ocli-ssh \
    --allow tcp:2222 \
    --source-ranges 0.0.0.0/0 \
    --target-tags ocli-ssh-server \
    --description "Allow OCLI SSH connections on port 2222" || echo "Firewall rule already exists"

# Copy binary to the instance
echo "📦 Uploading binary to instance..."
sleep 30  # Wait for instance to be ready
gcloud compute scp ocli-ssh-server $INSTANCE_NAME:/tmp/ --zone=$ZONE

# Move binary and set permissions
gcloud compute ssh $INSTANCE_NAME --zone=$ZONE --command="
    sudo mkdir -p /opt/ocli-ssh
    sudo mv /tmp/ocli-ssh-server /opt/ocli-ssh/
    sudo chmod +x /opt/ocli-ssh/ocli-ssh-server
    sudo chown -R ocli:ocli /opt/ocli-ssh
    sudo systemctl restart ocli-ssh
"

# Get external IP
EXTERNAL_IP=$(gcloud compute instances describe $INSTANCE_NAME --zone=$ZONE --format='get(networkInterfaces[0].accessConfigs[0].natIP)')

echo ""
echo "🎉 Deployment complete!"
echo "🌐 External IP: $EXTERNAL_IP"
echo "🔗 Connect with: ssh username@$EXTERNAL_IP -p 2222"
echo "📊 Check status: gcloud compute ssh $INSTANCE_NAME --zone=$ZONE --command='sudo systemctl status ocli-ssh'"
echo "📝 View logs: gcloud compute ssh $INSTANCE_NAME --zone=$ZONE --command='sudo journalctl -u ocli-ssh -f'"
echo ""
echo "💡 Don't forget to update PROJECT_ID in this script!"

# Cleanup
rm -f startup-script.sh ocli-ssh-server