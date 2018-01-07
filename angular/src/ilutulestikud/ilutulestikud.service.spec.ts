import { TestBed, inject } from '@angular/core/testing';

import { IlutulestikudService } from './ilutulestikud.service';

describe('IlutulestikudService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [IlutulestikudService]
    });
  });

  it('should be created', inject([IlutulestikudService], (service: IlutulestikudService) => {
    expect(service).toBeTruthy();
  }));
});
