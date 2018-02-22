import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { HttpClientModule } from '@angular/common/http';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MatSidenavModule } from '@angular/material/sidenav';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatSelectModule } from '@angular/material/select';
import { MatInputModule } from '@angular/material';
import { MatButtonModule } from '@angular/material/button';
import { MatDialogModule } from '@angular/material';
import { MatTabsModule } from '@angular/material/tabs';
import { MatListModule } from '@angular/material/list';
import { IlutulestikudComponent } from './ilutulestikud.component';
import { IlutulestikudService } from './ilutulestikud.service';
import { AddPlayerDialogueComponent } from './components/addplayerdialogue.component'
import { CreateGameDialogueComponent } from './components/creategamedialogue.component'
import { GameDisplayComponent } from './components/gamedisplay.component'

@NgModule({
  declarations: [
    IlutulestikudComponent,
    AddPlayerDialogueComponent,
    CreateGameDialogueComponent,
    GameDisplayComponent
  ],
  entryComponents: [
    AddPlayerDialogueComponent,
    CreateGameDialogueComponent
  ],
  imports: [
    BrowserModule,
    FormsModule,
    HttpClientModule,
    BrowserAnimationsModule,
    MatSidenavModule,
    MatToolbarModule,
    MatIconModule,
    MatCardModule,
    MatSelectModule,
    MatInputModule,
    MatButtonModule,
    MatDialogModule,
    MatTabsModule,
    MatListModule
  ],
  exports: [
    BrowserModule,
    FormsModule,
    HttpClientModule,
    BrowserAnimationsModule,
    MatSidenavModule,
    MatToolbarModule,
    MatIconModule,
    MatCardModule,
    MatSelectModule,
    MatInputModule,
    MatButtonModule,
    MatDialogModule,
    MatTabsModule,
    MatListModule,
    IlutulestikudComponent,
    GameDisplayComponent
  ],
  providers: [IlutulestikudService]
})
export class IlutulestikudModule { }
