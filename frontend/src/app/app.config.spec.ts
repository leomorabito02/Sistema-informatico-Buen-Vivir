import { appConfig } from './app.config';
import { ApplicationConfig } from '@angular/core';

describe('AppConfig', () => {
  it('should have providers configured', () => {
    expect(appConfig).toBeDefined();
    expect(appConfig.providers).toBeDefined();
    expect(Array.isArray(appConfig.providers)).toBeTruthy();
    expect(appConfig.providers.length).toBeGreaterThan(0);
  });
});
