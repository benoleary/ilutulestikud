import { CommonModule } from '@angular/common';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { InputTextModule } from 'primeng/primeng';

import { IlutulestikudComponent } from './ilutulestikud.component';
import { IlutulestikudService } from './ilutulestikud.service';


@NgModule({
  declarations: [
    IlutulestikudComponent
  ],
  imports: [
    CommonModule,
    FormsModule,
    HttpClientModule,
    InputTextModule
  ],
  exports: [
    IlutulestikudComponent
  ],
  providers: [IlutulestikudService]
})
export class IlutulestikudModule { }
