/**
 * Jest setup file for Chrome extension testing.
 * Sets up global mocks for Chrome APIs.
 */

// Mock Chrome APIs
const mockChrome = {
  runtime: {
    sendMessage: jest.fn().mockResolvedValue({ success: true }),
    onMessage: {
      addListener: jest.fn(),
      removeListener: jest.fn(),
    },
    sendNativeMessage: jest.fn().mockImplementation((_host, _msg, cb) => {
      if (cb) cb({ success: true });
    }),
    lastError: null as any,
    getURL: jest.fn().mockReturnValue('chrome-extension://abc/'),
  },
};

// Set global chrome variable
(globalThis as any).chrome = mockChrome;

// Set test flag
(globalThis as any).__TEST__ = true;

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    href: 'https://example.com',
    origin: 'https://example.com',
    pathname: '/',
    search: '',
    hash: '',
  },
  writable: true,
  configurable: true,
});

// Mock document
Object.defineProperty(document, 'title', {
  value: 'Test Page',
  writable: true,
  configurable: true,
});

// Mock document.readyState
let readyStateValue = 'complete';
Object.defineProperty(document, 'readyState', {
  get() { return readyStateValue; },
  set(val) { readyStateValue = val; },
  configurable: true,
});

// Mock MutationObserver
const mockMutationObserverObserve = jest.fn();
const mockMutationObserverDisconnect = jest.fn();

class MockMutationObserver {
  observe = mockMutationObserverObserve;
  disconnect = mockMutationObserverDisconnect;
  takeRecords = jest.fn();
}

(globalThis as any).MutationObserver = MockMutationObserver;
(globalThis as any).MockMutationObserver = {
  observe: mockMutationObserverObserve,
  disconnect: mockMutationObserverDisconnect,
};
