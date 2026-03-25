/**
 * Options Page Script
 * Handles cloud sync settings UI for the Verbalizer extension.
 */

interface AuthStatus {
  enabled: boolean;
  connected: boolean;
  email?: string;
  name?: string;
  provider?: string;
  scopes?: string;
}

interface Folder {
  id: string;
  name: string;
  parentId?: string;
}

interface NativeResponse {
  success: boolean;
  error?: string;
}

interface AuthStatusResponse {
  success: boolean;
  data?: AuthStatus;
  error?: string;
}

interface FolderResponse {
  success: boolean;
  data?: {
    folders: Folder[];
  };
  error?: string;
}

class OptionsPage {
  private bridge: NativeBridgeType;
  
  constructor() {
    // @ts-ignore - NativeBridge is injected by the build system
    this.bridge = window.nativeBridge;
    this.init();
  }
  
  private async init(): Promise<void> {
    await this.updateAuthStatus();
    this.setupEventListeners();
  }
  
  private setupEventListeners(): void {
    const btnConnect = document.getElementById('btnConnect');
    const btnDisconnect = document.getElementById('btnDisconnect');
    const btnSetFolder = document.getElementById('btnSetFolder');
    const folderSelect = document.getElementById('folderSelect') as HTMLSelectElement;
    
    btnConnect?.addEventListener('click', () => this.connectGoogle());
    btnDisconnect?.addEventListener('click', () => this.disconnectGoogle());
    btnSetFolder?.addEventListener('click', () => this.setFolder());
    folderSelect?.addEventListener('change', () => {
      (btnSetFolder as HTMLButtonElement).disabled = !folderSelect.value;
    });
  }
  
  private async updateAuthStatus(): Promise<void> {
    const statusEl = document.getElementById('authStatus');
    const emailRow = document.getElementById('emailRow');
    const emailEl = document.getElementById('accountEmail');
    const btnConnect = document.getElementById('btnConnect') as HTMLButtonElement;
    const btnDisconnect = document.getElementById('btnDisconnect') as HTMLButtonElement;
    const folderSection = document.getElementById('folderSection');
    
    try {
      const response = await this.bridge.googleAuthStatus();
      
      if (!response.success || !response.data) {
        this.showError(response.error || 'Failed to get auth status');
        return;
      }
      
      const status = response.data;
      
      if (!status.enabled) {
        statusEl!.textContent = 'Not configured';
        statusEl!.className = 'status-value status-disconnected';
        btnConnect!.style.display = 'none';
        btnDisconnect!.style.display = 'none';
        folderSection!.style.display = 'none';
        return;
      }
      
      if (status.connected) {
        statusEl!.textContent = 'Connected';
        statusEl!.className = 'status-value status-connected';
        emailRow!.style.display = 'flex';
        emailEl!.textContent = status.email || '-';
        btnConnect!.style.display = 'none';
        btnDisconnect!.style.display = 'inline-block';
        folderSection!.style.display = 'block';
        
        // Load folders
        await this.loadFolders();
      } else {
        statusEl!.textContent = 'Not connected';
        statusEl!.className = 'status-value status-disconnected';
        emailRow!.style.display = 'none';
        btnConnect!.style.display = 'inline-block';
        btnConnect!.disabled = false;
        btnDisconnect!.style.display = 'none';
        folderSection!.style.display = 'none';
      }
    } catch (err) {
      this.showError('Failed to connect to daemon');
    }
  }
  
  private async loadFolders(): Promise<void> {
    const folderSelect = document.getElementById('folderSelect') as HTMLSelectElement;
    
    try {
      folderSelect.innerHTML = '<option value="">Loading folders...</option>';
      
      const response = await this.bridge.googleDriveGetFolders();
      
      if (!response.success || !response.data) {
        folderSelect.innerHTML = '<option value="">Failed to load</option>';
        return;
      }
      
      const folders = response.data.folders || [];
      
      if (folders.length === 0) {
        folderSelect.innerHTML = '<option value="">No folders found</option>';
        return;
      }
      
      folderSelect.innerHTML = '<option value="">Select a folder...</option>';
      
      for (const folder of folders) {
        const option = document.createElement('option');
        option.value = folder.id;
        option.textContent = folder.name;
        folderSelect.appendChild(option);
      }
      
      // Enable selection
      folderSelect.disabled = false;
    } catch (err) {
      folderSelect.innerHTML = '<option value="">Error loading</option>';
    }
  }
  
  private async connectGoogle(): Promise<void> {
    const btnConnect = document.getElementById('btnConnect') as HTMLButtonElement;
    
    btnConnect.disabled = true;
    btnConnect.textContent = 'Connecting...';
    this.clearMessage();
    
    try {
      const response = await this.bridge.googleAuthStart();
      
      if (response.success) {
        this.showSuccess('Authentication started. A browser tab should have opened for Google sign-in.');
        // Poll for status change
        setTimeout(() => this.pollAuthStatus(), 2000);
      } else {
        this.showError(response.error || 'Failed to start authentication');
        btnConnect.disabled = false;
        btnConnect.textContent = 'Connect Google';
      }
    } catch (err) {
      this.showError('Failed to connect to daemon');
      btnConnect.disabled = false;
      btnConnect.textContent = 'Connect Google';
    }
  }
  
  private async pollAuthStatus(): Promise<void> {
    const maxAttempts = 30; // 30 * 2s = 60s timeout
    let attempts = 0;
    
    const poll = async () => {
      attempts++;
      await this.updateAuthStatus();
      
      const statusEl = document.getElementById('authStatus');
      if (statusEl?.textContent === 'Connected' || attempts >= maxAttempts) {
        return;
      }
      
      setTimeout(poll, 2000);
    };
    
    poll();
  }
  
  private async disconnectGoogle(): Promise<void> {
    const btnDisconnect = document.getElementById('btnDisconnect') as HTMLButtonElement;
    
    btnDisconnect.disabled = true;
    btnDisconnect.textContent = 'Disconnecting...';
    this.clearMessage();
    
    try {
      const response = await this.bridge.googleAuthDisconnect();
      
      if (response.success) {
        this.showSuccess('Google account disconnected.');
        await this.updateAuthStatus();
      } else {
        this.showError(response.error || 'Failed to disconnect');
      }
    } catch (err) {
      this.showError('Failed to connect to daemon');
    }
    
    btnDisconnect.disabled = false;
    btnDisconnect.textContent = 'Disconnect';
  }
  
  private async setFolder(): Promise<void> {
    const folderSelect = document.getElementById('folderSelect') as HTMLSelectElement;
    const btnSetFolder = document.getElementById('btnSetFolder') as HTMLButtonElement;
    
    if (!folderSelect.value) {
      this.showError('Please select a folder');
      return;
    }
    
    btnSetFolder.disabled = true;
    btnSetFolder.textContent = 'Setting...';
    this.clearMessage();
    
    try {
      const response = await this.bridge.googleDriveSetFolder(folderSelect.value);
      
      if (response.success) {
        this.showSuccess('Folder set successfully!');
        const folderEl = document.getElementById('targetFolder');
        const selectedOption = folderSelect.options[folderSelect.selectedIndex];
        if (folderEl && selectedOption) {
          folderEl.textContent = selectedOption.textContent;
        }
      } else {
        this.showError(response.error || 'Failed to set folder');
      }
    } catch (err) {
      this.showError('Failed to connect to daemon');
    }
    
    btnSetFolder.disabled = false;
    btnSetFolder.textContent = 'Set Folder';
  }
  
  private showError(message: string): void {
    const el = document.getElementById('messageArea');
    if (el) {
      el.innerHTML = `<p class="error">${message}</p>`;
    }
  }
  
  private showSuccess(message: string): void {
    const el = document.getElementById('messageArea');
    if (el) {
      el.innerHTML = `<p class="success">${message}</p>`;
    }
  }
  
  private clearMessage(): void {
    const el = document.getElementById('messageArea');
    if (el) {
      el.innerHTML = '';
    }
  }
}

// Native bridge type for TypeScript
interface NativeBridgeType {
  googleAuthStart(): Promise<NativeResponse>;
  googleAuthStatus(): Promise<AuthStatusResponse>;
  googleAuthDisconnect(): Promise<NativeResponse>;
  googleDriveGetFolders(): Promise<FolderResponse>;
  googleDriveSetFolder(folderId: string): Promise<NativeResponse>;
}

// Initialize when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  new OptionsPage();
});