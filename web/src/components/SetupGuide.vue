<script setup lang="ts">
import { ref } from 'vue'
import { getConnectionStatus } from '../api/client'

const copied = ref<string | null>(null)

async function copyToClipboard(text: string, id: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = id
    setTimeout(() => {
      copied.value = null
    }, 2000)
  } catch {
    // Clipboard not available
  }
}

async function checkConnection() {
  try {
    await getConnectionStatus()
    window.location.href = '/'
  } catch {
    // Connection still failing
  }
}
</script>

<template>
  <div class="setup-guide">
    <header class="guide-header">
      <h1>AWS Setup Guide</h1>
      <p class="subtitle">
        WeaveLens uses credentials available to the backend runtime environment.
        Follow the instructions below to configure AWS access.
      </p>
    </header>

    <div class="guide-content">
      <section class="setup-section">
        <h2>Local Development</h2>
        <ol class="steps">
          <li>
            <strong>Install and configure the AWS CLI</strong>
            <p>If you haven't already, install the AWS CLI and configure a profile:</p>
            <div class="code-block">
              <code>aws configure --profile weavelens</code>
              <button @click="copyToClipboard('aws configure --profile weavelens', 'cmd1')" class="copy-btn">
                {{ copied === 'cmd1' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
          </li>
          <li>
            <strong>Start WeaveLens with the profile</strong>
            <p>Set the <code>AWS_PROFILE</code> environment variable when starting WeaveLens:</p>
            <p><em>Option A: Run from source</em></p>
            <div class="code-block">
              <code>AWS_PROFILE=weavelens AWS_REGION=us-east-1 go run ./cmd/weavelens</code>
              <button @click="copyToClipboard('AWS_PROFILE=weavelens AWS_REGION=us-east-1 go run ./cmd/weavelens', 'cmd2')" class="copy-btn">
                {{ copied === 'cmd2' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <p><em>Option B: Build and run binary</em></p>
            <div class="code-block">
              <code>go build -o weavelens ./cmd/weavelens && AWS_PROFILE=weavelens AWS_REGION=us-east-1 ./weavelens</code>
              <button @click="copyToClipboard('go build -o weavelens ./cmd/weavelens\nAWS_PROFILE=weavelens AWS_REGION=us-east-1 ./weavelens', 'cmd2b')" class="copy-btn">
                {{ copied === 'cmd2b' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
            <p><em>Or export variables first</em></p>
            <div class="code-block">
              <code>export AWS_PROFILE=weavelens&#10;export AWS_REGION=us-east-1&#10;go run ./cmd/weavelens</code>
              <button @click="copyToClipboard('export AWS_PROFILE=weavelens\nexport AWS_REGION=us-east-1\ngo run ./cmd/weavelens', 'cmd2c')" class="copy-btn">
                {{ copied === 'cmd2c' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
          </li>
          <li>
            <strong>Verify the connection</strong>
            <p>Return to WeaveLens and check the connection status.</p>
          </li>
        </ol>
      </section>

      <section class="setup-section">
        <h2>Environment Credentials</h2>
        <p>
          Alternatively, you can use environment variables:
        </p>
        <div class="code-block">
          <code>AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_REGION=us-east-1 go run ./cmd/weavelens</code>
          <button @click="copyToClipboard('AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_REGION=us-east-1 go run ./cmd/weavelens', 'cmd3')" class="copy-btn">
            {{ copied === 'cmd3' ? '✓ Copied' : 'Copy' }}
          </button>
        </div>
        <p>Or after building:</p>
        <div class="code-block">
          <code>AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_REGION=us-east-1 ./weavelens</code>
          <button @click="copyToClipboard('AWS_ACCESS_KEY_ID=... AWS_SECRET_ACCESS_KEY=... AWS_REGION=us-east-1 ./weavelens', 'cmd3b')" class="copy-btn">
            {{ copied === 'cmd3b' ? '✓ Copied' : 'Copy' }}
          </button>
        </div>
        <p class="warning">
          ⚠ Environment variables are suitable for CI/CD pipelines but should be
          managed through secret stores in production.
        </p>
      </section>

      <section class="setup-section">
        <h2>LocalStack (Local AWS Emulator)</h2>
        <p>
          For local development without real AWS credentials, use LocalStack:
        </p>
        <div class="code-block">
          <code>docker run -d --name localstack -p 4566:4566 localstack/localstack</code>
          <button @click="copyToClipboard('docker run -d --name localstack -p 4566:4566 localstack/localstack', 'ls1')" class="copy-btn">
            {{ copied === 'ls1' ? '✓ Copied' : 'Copy' }}
          </button>
        </div>
        <p>Then configure WeaveLens to use LocalStack:</p>
        <div class="code-block">
          <code>export AWS_ENDPOINT_URL=http://localhost:4566&#10;export AWS_ACCESS_KEY_ID=test&#10;export AWS_SECRET_ACCESS_KEY=test&#10;export AWS_REGION=us-east-1&#10;go run ./cmd/weavelens</code>
          <button @click="copyToClipboard('export AWS_ENDPOINT_URL=http://localhost:4566\nexport AWS_ACCESS_KEY_ID=test\nexport AWS_SECRET_ACCESS_KEY=test\nexport AWS_REGION=us-east-1\ngo run ./cmd/weavelens', 'ls2')" class="copy-btn">
            {{ copied === 'ls2' ? '✓ Copied' : 'Copy' }}
          </button>
        </div>
        <p class="note">
          LocalStack accepts any credential values. The endpoint URL is required to redirect AWS API calls to LocalStack.
        </p>
      </section>

      <section class="setup-section">
        <h2>Cross-Account Access (STS AssumeRole)</h2>
        <p>
          WeaveLens can use IAM roles for cross-account access. The runtime identity
          must have permission to assume the target role.
        </p>
        <div class="code-block">
          <code>AWS_ROLE_ARN=arn:aws:iam::123456789012:role/WeaveLensScanner go run ./cmd/weavelens</code>
          <button @click="copyToClipboard('AWS_ROLE_ARN=arn:aws:iam::123456789012:role/WeaveLensScanner go run ./cmd/weavelens', 'cmd4')" class="copy-btn">
            {{ copied === 'cmd4' ? '✓ Copied' : 'Copy' }}
          </button>
        </div>
        <p>Or after building:</p>
        <div class="code-block">
          <code>AWS_ROLE_ARN=arn:aws:iam::123456789012:role/WeaveLensScanner ./weavelens</code>
          <button @click="copyToClipboard('AWS_ROLE_ARN=arn:aws:iam::123456789012:role/WeaveLensScanner ./weavelens', 'cmd4b')" class="copy-btn">
            {{ copied === 'cmd4b' ? '✓ Copied' : 'Copy' }}
          </button>
        </div>
        <p class="note">
          The target role must grant the required permissions for resource discovery
          and must trust the runtime identity.
        </p>
      </section>

      <section class="setup-section">
        <h2>Required IAM Permissions</h2>
        <p>
          The AWS identity used by WeaveLens needs the following permissions:
        </p>
        <div class="permissions-grid">
          <div class="permission-category">
            <h3>EC2</h3>
            <ul>
              <li>DescribeVpcs</li>
              <li>DescribeSubnets</li>
              <li>DescribeRouteTables</li>
              <li>DescribeInternetGateways</li>
              <li>DescribeNatGateways</li>
              <li>DescribeSecurityGroups</li>
              <li>DescribeInstances</li>
            </ul>
          </div>
          <div class="permission-category">
            <h3>RDS</h3>
            <ul>
              <li>DescribeDBInstances</li>
            </ul>
          </div>
          <div class="permission-category">
            <h3>ELB</h3>
            <ul>
              <li>DescribeLoadBalancers</li>
            </ul>
          </div>
          <div class="permission-category">
            <h3>STS</h3>
            <ul>
              <li>GetCallerIdentity</li>
            </ul>
          </div>
        </div>
      </section>

      <section class="setup-section">
        <h2>Troubleshooting</h2>
        <div class="troubleshooting">
          <div class="trouble-item">
            <h3>Authentication Error</h3>
            <p>Verify your credentials are valid and not expired:</p>
            <div class="code-block">
              <code>aws sts get-caller-identity</code>
              <button @click="copyToClipboard('aws sts get-caller-identity', 'cmd5')" class="copy-btn">
                {{ copied === 'cmd5' ? '✓ Copied' : 'Copy' }}
              </button>
            </div>
          </div>
          <div class="trouble-item">
            <h3>Access Denied</h3>
            <p>Ensure the IAM identity has the required permissions listed above.</p>
          </div>
          <div class="trouble-item">
            <h3>Region Not Configured</h3>
            <p>Set the <code>AWS_REGION</code> environment variable or use <code>aws configure</code> to set a default region.</p>
          </div>
        </div>
      </section>
    </div>

    <footer class="guide-footer">
      <button @click="checkConnection" class="primary-btn">Check Connection</button>
      <router-link to="/" class="secondary-btn">Back to Dashboard</router-link>
    </footer>
  </div>
</template>

<style scoped>
.setup-guide {
  max-width: 900px;
  margin: 0 auto;
  padding: 32px;
  min-height: 100vh;
}

.guide-header {
  margin-bottom: 32px;
  padding-bottom: 24px;
  border-bottom: 2px solid #e0e0e0;
}

.guide-header h1 {
  margin: 0 0 8px 0;
  font-size: 28px;
  color: #333;
}

.subtitle {
  color: #666;
  font-size: 15px;
  line-height: 1.5;
  max-width: 600px;
}

.guide-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.setup-section {
  background: #fafafa;
  border-radius: 8px;
  padding: 20px;
}

.setup-section h2 {
  margin: 0 0 16px 0;
  font-size: 18px;
  color: #333;
}

.steps {
  padding-left: 24px;
  margin: 0;
}

.steps li {
  margin-bottom: 16px;
}

.steps li strong {
  display: block;
  margin-bottom: 4px;
  color: #333;
}

.steps li p {
  margin: 4px 0;
  color: #666;
  font-size: 14px;
}

.code-block {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #263238;
  color: #e0e0e0;
  padding: 12px 16px;
  border-radius: 6px;
  margin: 8px 0;
  gap: 12px;
}

.code-block code {
  font-family: monospace;
  font-size: 13px;
  word-break: break-all;
  flex: 1;
  white-space: pre-wrap;
}

.copy-btn {
  padding: 6px 12px;
  background: #37474f;
  color: #e0e0e0;
  border: 1px solid #455a64;
  border-radius: 4px;
  cursor: pointer;
  font-size: 11px;
  white-space: nowrap;
  flex-shrink: 0;
}

.copy-btn:hover {
  background: #455a64;
}

.warning {
  color: #e65100;
  font-size: 13px;
  margin-top: 12px;
  padding: 8px 12px;
  background: #fff3e0;
  border-radius: 4px;
}

.note {
  color: #666;
  font-size: 13px;
  margin-top: 8px;
}

.permissions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.permission-category {
  background: white;
  padding: 12px;
  border-radius: 6px;
  border: 1px solid #e0e0e0;
}

.permission-category h3 {
  margin: 0 0 8px 0;
  font-size: 13px;
  color: #1976d2;
}

.permission-category ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.permission-category li {
  font-size: 12px;
  color: #666;
  padding: 2px 0;
}

.troubleshooting {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.trouble-item h3 {
  margin: 0 0 4px 0;
  font-size: 14px;
  color: #333;
}

.trouble-item p {
  margin: 4px 0;
  color: #666;
  font-size: 13px;
}

.guide-footer {
  display: flex;
  gap: 12px;
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #e0e0e0;
}

.primary-btn {
  padding: 10px 24px;
  background: #1976d2;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
}

.primary-btn:hover {
  background: #1565c0;
}

.secondary-btn {
  padding: 10px 24px;
  background: transparent;
  color: #1976d2;
  border: 1px solid #1976d2;
  border-radius: 6px;
  cursor: pointer;
  font-size: 14px;
  text-decoration: none;
  display: flex;
  align-items: center;
}

.secondary-btn:hover {
  background: #e3f2fd;
}

/* Responsive */
@media (max-width: 768px) {
  .setup-guide {
    padding: 16px;
  }

  .guide-header h1 {
    font-size: 22px;
  }

  .permissions-grid {
    grid-template-columns: 1fr 1fr;
  }

  .code-block {
    flex-direction: column;
    align-items: stretch;
  }

  .copy-btn {
    align-self: flex-end;
  }
}
</style>
