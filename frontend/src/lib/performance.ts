// Performance monitoring utilities for the frontend

interface PerformanceMetrics {
  navigationTiming: PerformanceNavigationTiming | null;
  resourceTimings: PerformanceResourceTiming[];
  paintTimings: PerformanceEntry[];
  customMetrics: Map<string, number>;
}

interface WebVitals {
  CLS: number | null; // Cumulative Layout Shift
  FID: number | null; // First Input Delay
  FCP: number | null; // First Contentful Paint
  LCP: number | null; // Largest Contentful Paint
  TTFB: number | null; // Time to First Byte
}

class PerformanceMonitor {
  private metrics: PerformanceMetrics;
  private webVitals: WebVitals;
  private observers: Map<string, PerformanceObserver>;

  constructor() {
    this.metrics = {
      navigationTiming: null,
      resourceTimings: [],
      paintTimings: [],
      customMetrics: new Map(),
    };

    this.webVitals = {
      CLS: null,
      FID: null,
      FCP: null,
      LCP: null,
      TTFB: null,
    };

    this.observers = new Map();
    this.initializeObservers();
  }

  private initializeObservers(): void {
    if (typeof window === 'undefined' || !('PerformanceObserver' in window)) {
      return;
    }

    // Observe navigation timing
    this.observeNavigationTiming();

    // Observe resource timing
    this.observeResourceTiming();

    // Observe paint timing
    this.observePaintTiming();

    // Observe layout shift (CLS)
    this.observeLayoutShift();

    // Observe first input delay (FID)
    this.observeFirstInputDelay();

    // Observe largest contentful paint (LCP)
    this.observeLargestContentfulPaint();
  }

  private observeNavigationTiming(): void {
    try {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        for (const entry of entries) {
          if (entry.entryType === 'navigation') {
            this.metrics.navigationTiming = entry as PerformanceNavigationTiming;
            this.calculateTTFB();
          }
        }
      });

      observer.observe({ entryTypes: ['navigation'] });
      this.observers.set('navigation', observer);
    } catch (error) {
      console.warn('Navigation timing observer not supported:', error);
    }
  }

  private observeResourceTiming(): void {
    try {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries() as PerformanceResourceTiming[];
        this.metrics.resourceTimings.push(...entries);
      });

      observer.observe({ entryTypes: ['resource'] });
      this.observers.set('resource', observer);
    } catch (error) {
      console.warn('Resource timing observer not supported:', error);
    }
  }

  private observePaintTiming(): void {
    try {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        this.metrics.paintTimings.push(...entries);
        
        for (const entry of entries) {
          if (entry.name === 'first-contentful-paint') {
            this.webVitals.FCP = entry.startTime;
          }
        }
      });

      observer.observe({ entryTypes: ['paint'] });
      this.observers.set('paint', observer);
    } catch (error) {
      console.warn('Paint timing observer not supported:', error);
    }
  }

  private observeLayoutShift(): void {
    try {
      let clsValue = 0;
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          if (!(entry as any).hadRecentInput) {
            clsValue += (entry as any).value;
            this.webVitals.CLS = clsValue;
          }
        }
      });

      observer.observe({ entryTypes: ['layout-shift'] });
      this.observers.set('layout-shift', observer);
    } catch (error) {
      console.warn('Layout shift observer not supported:', error);
    }
  }

  private observeFirstInputDelay(): void {
    try {
      const observer = new PerformanceObserver((list) => {
        for (const entry of list.getEntries()) {
          this.webVitals.FID = (entry as any).processingStart - entry.startTime;
        }
      });

      observer.observe({ entryTypes: ['first-input'] });
      this.observers.set('first-input', observer);
    } catch (error) {
      console.warn('First input delay observer not supported:', error);
    }
  }

  private observeLargestContentfulPaint(): void {
    try {
      const observer = new PerformanceObserver((list) => {
        const entries = list.getEntries();
        const lastEntry = entries[entries.length - 1];
        this.webVitals.LCP = lastEntry.startTime;
      });

      observer.observe({ entryTypes: ['largest-contentful-paint'] });
      this.observers.set('largest-contentful-paint', observer);
    } catch (error) {
      console.warn('Largest contentful paint observer not supported:', error);
    }
  }

  private calculateTTFB(): void {
    if (this.metrics.navigationTiming) {
      this.webVitals.TTFB = this.metrics.navigationTiming.responseStart - 
                           this.metrics.navigationTiming.requestStart;
    }
  }

  // Public methods
  public markStart(name: string): void {
    if (typeof window !== 'undefined' && 'performance' in window) {
      performance.mark(`${name}-start`);
    }
  }

  public markEnd(name: string): void {
    if (typeof window !== 'undefined' && 'performance' in window) {
      performance.mark(`${name}-end`);
      performance.measure(name, `${name}-start`, `${name}-end`);
      
      const measure = performance.getEntriesByName(name, 'measure')[0];
      if (measure) {
        this.metrics.customMetrics.set(name, measure.duration);
      }
    }
  }

  public recordCustomMetric(name: string, value: number): void {
    this.metrics.customMetrics.set(name, value);
  }

  public getMetrics(): PerformanceMetrics {
    return { ...this.metrics };
  }

  public getWebVitals(): WebVitals {
    return { ...this.webVitals };
  }

  public getResourcesByType(type: string): PerformanceResourceTiming[] {
    return this.metrics.resourceTimings.filter(resource => 
      resource.initiatorType === type
    );
  }

  public getLargestResources(count: number = 10): PerformanceResourceTiming[] {
    return this.metrics.resourceTimings
      .sort((a, b) => b.transferSize - a.transferSize)
      .slice(0, count);
  }

  public getSlowestResources(count: number = 10): PerformanceResourceTiming[] {
    return this.metrics.resourceTimings
      .sort((a, b) => (b.responseEnd - b.requestStart) - (a.responseEnd - a.requestStart))
      .slice(0, count);
  }

  public generateReport(): string {
    const vitals = this.getWebVitals();
    const metrics = this.getMetrics();
    
    let report = '=== Performance Report ===\n\n';
    
    // Web Vitals
    report += 'Web Vitals:\n';
    report += `  CLS: ${vitals.CLS?.toFixed(3) || 'N/A'}\n`;
    report += `  FID: ${vitals.FID?.toFixed(2) || 'N/A'}ms\n`;
    report += `  FCP: ${vitals.FCP?.toFixed(2) || 'N/A'}ms\n`;
    report += `  LCP: ${vitals.LCP?.toFixed(2) || 'N/A'}ms\n`;
    report += `  TTFB: ${vitals.TTFB?.toFixed(2) || 'N/A'}ms\n\n`;
    
    // Navigation Timing
    if (metrics.navigationTiming) {
      const nav = metrics.navigationTiming;
      report += 'Navigation Timing:\n';
      report += `  DNS Lookup: ${(nav.domainLookupEnd - nav.domainLookupStart).toFixed(2)}ms\n`;
      report += `  TCP Connect: ${(nav.connectEnd - nav.connectStart).toFixed(2)}ms\n`;
      report += `  Request: ${(nav.responseStart - nav.requestStart).toFixed(2)}ms\n`;
      report += `  Response: ${(nav.responseEnd - nav.responseStart).toFixed(2)}ms\n`;
      report += `  DOM Processing: ${(nav.domComplete - nav.domInteractive).toFixed(2)}ms\n`;
      report += `  Load Complete: ${(nav.loadEventEnd - nav.loadEventStart).toFixed(2)}ms\n\n`;
    }
    
    // Custom Metrics
    if (metrics.customMetrics.size > 0) {
      report += 'Custom Metrics:\n';
      metrics.customMetrics.forEach((value, name) => {
        report += `  ${name}: ${value.toFixed(2)}ms\n`;
      });
      report += '\n';
    }
    
    // Resource Summary
    const totalResources = metrics.resourceTimings.length;
    const totalSize = metrics.resourceTimings.reduce((sum, resource) => 
      sum + (resource.transferSize || 0), 0
    );
    
    report += 'Resource Summary:\n';
    report += `  Total Resources: ${totalResources}\n`;
    report += `  Total Size: ${(totalSize / 1024).toFixed(2)} KB\n`;
    
    return report;
  }

  public sendToAnalytics(endpoint?: string): void {
    const data = {
      webVitals: this.getWebVitals(),
      customMetrics: Object.fromEntries(this.metrics.customMetrics),
      timestamp: Date.now(),
      userAgent: navigator.userAgent,
      url: window.location.href,
    };

    if (endpoint) {
      fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      }).catch(error => {
        console.warn('Failed to send performance data:', error);
      });
    }
  }

  public disconnect(): void {
    this.observers.forEach((observer) => {
      observer.disconnect();
    });
    this.observers.clear();
  }
}

// Singleton instance
let performanceMonitor: PerformanceMonitor | null = null;

export function getPerformanceMonitor(): PerformanceMonitor {
  if (!performanceMonitor) {
    performanceMonitor = new PerformanceMonitor();
  }
  return performanceMonitor;
}

// React hook for performance monitoring
export function usePerformanceMonitor() {
  const monitor = getPerformanceMonitor();
  
  return {
    markStart: monitor.markStart.bind(monitor),
    markEnd: monitor.markEnd.bind(monitor),
    recordMetric: monitor.recordCustomMetric.bind(monitor),
    getMetrics: monitor.getMetrics.bind(monitor),
    getWebVitals: monitor.getWebVitals.bind(monitor),
    generateReport: monitor.generateReport.bind(monitor),
  };
}

// Utility functions
export function measureAsyncOperation<T>(
  name: string,
  operation: () => Promise<T>
): Promise<T> {
  const monitor = getPerformanceMonitor();
  monitor.markStart(name);
  
  return operation().finally(() => {
    monitor.markEnd(name);
  });
}

export function measureSyncOperation<T>(
  name: string,
  operation: () => T
): T {
  const monitor = getPerformanceMonitor();
  monitor.markStart(name);
  
  try {
    return operation();
  } finally {
    monitor.markEnd(name);
  }
}

// Initialize performance monitoring on module load
if (typeof window !== 'undefined') {
  getPerformanceMonitor();
}
