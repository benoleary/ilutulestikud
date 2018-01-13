import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MatSelectModule } from '@angular/material/select';
import { IlutulestikudComponent } from './ilutulestikud.component';
import { IlutulestikudService } from './ilutulestikud.service';


@NgModule({
  declarations: [
    IlutulestikudComponent
  ],
  imports: [
    BrowserModule,
    FormsModule,
    HttpClientModule,
    BrowserAnimationsModule,
    MatSelectModule
  ],
  exports: [
    BrowserModule,
    FormsModule,
    HttpClientModule,
    BrowserAnimationsModule,
    MatSelectModule,
    IlutulestikudComponent
  ],
  providers: [IlutulestikudService]
})
export class IlutulestikudModule { }
