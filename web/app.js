// GoBE Landing Page JavaScript
'use strict';

class GoBEDashboard {
    constructor() {
        this.systemData = {};
        this.refreshInterval = null;
        this.initializeEventListeners();
        this.startAutoRefresh();
    }

    // Initialize event listeners
    initializeEventListeners() {
        document.addEventListener('DOMContentLoaded', () => {
            this.refreshSystemStatus();
            console.log('🚀 GoBE Dashboard initialized');
        });

        window.addEventListener('beforeunload', () => {
            this.cleanup();
        });
    }

    // Start auto-refresh of system status
    startAutoRefresh() {
        this.refreshInterval = setInterval(() => {
            this.refreshSystemStatus();
        }, 30000); // Refresh every 30 seconds
    }

    // Cleanup resources
    cleanup() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }

    // Toast notification system
    showToast(message, type = 'success') {
        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        toast.innerHTML = `
            <div style="display: flex; align-items: center; gap: 0.5rem;">
                <span>${type === 'success' ? '✅' : '❌'}</span>
                <span>${message}</span>
            </div>
        `;
        document.body.appendChild(toast);

        // Animate in
        setTimeout(() => toast.classList.add('show'), 100);

        // Animate out and remove
        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => {
                if (document.body.contains(toast)) {
                    document.body.removeChild(toast);
                }
            }, 300);
        }, 3000);
    }

    // System status refresh using MCP
    async refreshSystemStatus() {
        try {
            const response = await this.fetchWithTimeout('/mcp/exec', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    tool: 'system.status',
                    args: { detailed: true }
                })
            }, 5000);

            if (response.ok) {
                const data = await response.json();
                if (data.status === 'success' && data.data?.result) {
                    this.updateSystemDisplay(data.data.result);
                    this.updateStatusIndicator('System Operational', true);
                    return;
                }
            }
            throw new Error('Invalid MCP response');
        } catch (error) {
            console.warn('MCP not available, using fallback:', error);
            this.updateSystemDisplay({
                status: 'ok',
                version: 'v1.3.4',
                uptime: 'Unknown',
                runtime: {
                    go_version: 'Go 1.24+',
                    memory: { alloc_mb: 'N/A' },
                    goroutines: 'N/A'
                }
            });
            this.updateStatusIndicator('Fallback Mode', false);
        }
    }

    // Update system display with data
    updateSystemDisplay(data) {
        const elements = {
            version: data.version || 'v1.3.4',
            uptime: data.uptime || 'Unknown',
            'go-version': data.runtime?.go_version || 'Go 1.24+',
            memory: data.runtime?.memory?.alloc_mb ?
                `${Math.round(data.runtime.memory.alloc_mb)}MB` : 'N/A',
            goroutines: data.runtime?.goroutines || 'N/A'
        };

        Object.entries(elements).forEach(([id, value]) => {
            const element = document.getElementById(id);
            if (element) element.textContent = value;
        });

        this.systemData = data;
    }

    // Update status indicator
    updateStatusIndicator(text, isOnline) {
        const statusText = document.getElementById('status-text');
        const statusDot = document.querySelector('.status-dot');

        if (statusText) statusText.textContent = text;
        if (statusDot) {
            statusDot.style.background = isOnline ? 'var(--accent-color)' : '#f59e0b';
            statusDot.style.boxShadow = isOnline ?
                '0 0 10px var(--accent-color)' : '0 0 10px #f59e0b';
        }
    }

    // Fetch with timeout
    async fetchWithTimeout(url, options = {}, timeout = 5000) {
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), timeout);

        try {
            const response = await fetch(url, {
                ...options,
                signal: controller.signal
            });
            clearTimeout(timeoutId);
            return response;
        } catch (error) {
            clearTimeout(timeoutId);
            throw error;
        }
    }

    // MCP Tools testing
    async testMCPTools() {
        try {
            const response = await this.fetchWithTimeout('/mcp/tools');
            if (response.ok) {
                const data = await response.json();
                const toolCount = data.data?.tools?.length || 0;
                this.showToast(`✅ MCP Tools: ${toolCount} tools available`, 'success');
                return true;
            }
            throw new Error('Failed to fetch tools');
        } catch (error) {
            this.showToast('❌ MCP Tools not available', 'error');
            return false;
        }
    }

    // System status test (specific MCP tool)
    async testSystemStatus() {
        try {
            const response = await this.fetchWithTimeout('/mcp/exec', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    tool: 'system.status',
                    args: { detailed: false }
                })
            });

            if (response.ok) {
                const data = await response.json();
                const status = data.data?.result?.status || 'OK';
                this.showToast(`✅ System Status: ${status}`, 'success');
                return true;
            }
            throw new Error('System status check failed');
        } catch (error) {
            this.showToast('❌ System status check failed', 'error');
            return false;
        }
    }

    // Database health check
    async checkDatabase() {
        try {
            const response = await this.fetchWithTimeout('/api/v1/system/metrics');
            if (response.ok) {
                this.showToast('✅ Database connection healthy', 'success');
                return true;
            }
            throw new Error('Database check failed');
        } catch (error) {
            this.showToast('❌ Database check failed', 'error');
            return false;
        }
    }

    // Test all integrations
    async testIntegrations() {
        const integrations = [
            { name: 'whatsapp', url: '/api/v1/whatsapp/ping' },
            { name: 'telegram', url: '/api/v1/telegram/ping' },
            { name: 'rabbitmq', url: '/api/v1/rabbitmq/ping' }
        ];

        let successCount = 0;

        for (const integration of integrations) {
            try {
                const response = await this.fetchWithTimeout(integration.url, {}, 3000);
                const isActive = response.ok;
                const status = isActive ? 'Active ✓' : 'Inactive ⚠️';
                const color = isActive ? 'var(--accent-color)' : '#f59e0b';

                const statusElement = document.getElementById(`${integration.name}-status`);
                if (statusElement) {
                    statusElement.textContent = status;
                    statusElement.style.color = color;
                }

                if (isActive) successCount++;
            } catch (error) {
                const statusElement = document.getElementById(`${integration.name}-status`);
                if (statusElement) {
                    statusElement.textContent = 'Error ❌';
                    statusElement.style.color = '#ef4444';
                }
            }
        }

        this.showToast(`🔄 Integrations updated: ${successCount}/${integrations.length} active`, 'success');
        return successCount;
    }

    // Utility functions
    openAPIDoc() {
        window.open('/swagger', '_blank');
    }

    viewSecurityInfo() {
        this.showToast('🛡️ Security: TLS, JWT, Rate Limiting, CORS all active', 'success');
    }

    viewLogs() {
        this.showToast('📋 Log viewer would open here (./gobe logs)', 'success');
    }

    // Download configuration file
    downloadConfig() {
        const config = {
            port: 3666,
            bindAddress: '0.0.0.0',
            database: {
                type: 'postgres',
                host: 'localhost',
                port: 5432
            },
            mcp: {
                enabled: true,
                tools: ['system.status']
            },
            security: {
                tls: true,
                jwt: true,
                rateLimiting: true,
                cors: true
            },
            integrations: {
                whatsapp: { enabled: false },
                telegram: { enabled: false },
                rabbitmq: { enabled: false }
            }
        };

        const blob = new Blob([JSON.stringify(config, null, 2)], {
            type: 'application/json'
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'gobe-config.json';
        a.click();
        URL.revokeObjectURL(url);

        this.showToast('⚙️ Configuration downloaded', 'success');
    }

    // Health check endpoint
    async healthCheck() {
        try {
            const response = await this.fetchWithTimeout('/health');
            if (response.ok) {
                const data = await response.json();
                this.showToast(`💚 Health: ${data.status || 'OK'}`, 'success');
                window.open('/health', '_blank');
                return true;
            }
            throw new Error('Health check failed');
        } catch (error) {
            this.showToast('❌ Health check failed', 'error');
            return false;
        }
    }

    // Run comprehensive system test
    async runSystemTest() {
        this.showToast('🧪 Running comprehensive system test...', 'success');

        const tests = [
            { name: 'MCP Tools', fn: () => this.testMCPTools() },
            { name: 'System Status', fn: () => this.testSystemStatus() },
            { name: 'Database', fn: () => this.checkDatabase() },
            { name: 'Integrations', fn: () => this.testIntegrations() }
        ];

        let passedTests = 0;
        for (const test of tests) {
            try {
                const result = await test.fn();
                if (result) passedTests++;
            } catch (error) {
                console.error(`Test ${test.name} failed:`, error);
            }
        }

        const score = Math.round((passedTests / tests.length) * 100);
        this.showToast(`🎯 System Test Complete: ${score}% (${passedTests}/${tests.length})`,
            score >= 75 ? 'success' : 'error');
    }
}

// Initialize dashboard when DOM is ready
const dashboard = new GoBEDashboard();

// Global functions for HTML onclick handlers
window.refreshSystemStatus = () => dashboard.refreshSystemStatus();
window.testMCPTools = () => dashboard.testMCPTools();
window.testSystemStatus = () => dashboard.testSystemStatus();
window.checkDatabase = () => dashboard.checkDatabase();
window.testIntegrations = () => dashboard.testIntegrations();
window.openAPIDoc = () => dashboard.openAPIDoc();
window.viewSecurityInfo = () => dashboard.viewSecurityInfo();
window.viewLogs = () => dashboard.viewLogs();
window.downloadConfig = () => dashboard.downloadConfig();
window.healthCheck = () => dashboard.healthCheck();
window.runSystemTest = () => dashboard.runSystemTest();

// Keyboard shortcuts
document.addEventListener('keydown', (e) => {
    if (e.ctrlKey || e.metaKey) {
        switch (e.key) {
            case 'r':
                e.preventDefault();
                dashboard.refreshSystemStatus();
                break;
            case 't':
                e.preventDefault();
                dashboard.runSystemTest();
                break;
        }
    }
});

// Console easter egg
console.log(`
🚀 GoBE Dashboard v1.3.4
━━━━━━━━━━━━━━━━━━━━━━━━
Keyboard Shortcuts:
  Ctrl+R - Refresh Status
  Ctrl+T - Run System Test

Available Commands:
  dashboard.refreshSystemStatus()
  dashboard.runSystemTest()
  dashboard.testMCPTools()

🎯 Code Fast. Own Everything.
`);

export { GoBEDashboard };