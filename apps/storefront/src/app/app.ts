import { CommonModule } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Component, OnInit, computed, inject, signal } from '@angular/core';

interface Product { id:string; name:string; description:string; priceCents:number; currency:string; imageUrl:string; inStock:boolean }
interface CartItem { productId:string; quantity:number }
interface Cart { customerId:string; items:CartItem[] }
interface Order { id:string; customerId:string; status:string; currency:string; totalCents:number; items:Array<CartItem & {unitPriceCents:number}> }

@Component({selector:'app-root',imports:[CommonModule],templateUrl:'./app.html',styleUrl:'./app.css'})
export class App implements OnInit {
  private readonly http=inject(HttpClient); readonly customerId='customer-storefront-demo';
  readonly products=signal<Product[]>([]); readonly cart=signal<Cart>({customerId:this.customerId,items:[]}); readonly order=signal<Order|null>(null);
  readonly loading=signal(true); readonly busy=signal(false); readonly error=signal('');
  readonly cartCount=computed(()=>this.cart().items.reduce((sum,item)=>sum+item.quantity,0));
  readonly cartTotal=computed(()=>this.cart().items.reduce((sum,item)=>sum+item.quantity*(this.product(item.productId)?.priceCents??0),0));
  private get headers(){return new HttpHeaders({'X-Customer-ID':this.customerId})}
  ngOnInit(){this.loadProducts();this.loadCart()}
  product(id:string){return this.products().find(value=>value.id===id)}
  money(cents:number){return new Intl.NumberFormat('en-US',{style:'currency',currency:'USD'}).format(cents/100)}
  loadProducts(){this.loading.set(true);this.http.get<Product[]>('/api/products').subscribe({next:value=>{this.products.set(value);this.loading.set(false)},error:()=>{this.error.set('Could not load the catalog.');this.loading.set(false)}})}
  loadCart(){this.http.get<Cart>('/api/cart',{headers:this.headers}).subscribe({next:value=>this.cart.set(value),error:()=>this.error.set('Could not load your cart.')})}
  add(product:Product){this.busy.set(true);this.http.post('/api/cart/items',{productId:product.id,quantity:1},{headers:this.headers}).subscribe({next:()=>{this.loadCart();this.busy.set(false)},error:()=>{this.error.set('Could not add the item.');this.busy.set(false)}})}
  change(item:CartItem,delta:number){const quantity=item.quantity+delta;if(quantity<=0){this.remove(item);return}this.http.patch(`/api/cart/items/${item.productId}`,{quantity},{headers:this.headers}).subscribe(()=>this.loadCart())}
  remove(item:CartItem){this.http.delete(`/api/cart/items/${item.productId}`,{headers:this.headers}).subscribe(()=>this.loadCart())}
  checkout(){const items=this.cart().items.map(item=>({...item,unitPriceCents:this.product(item.productId)?.priceCents??0}));if(!items.length)return;this.busy.set(true);this.error.set('');this.http.post<Order>('/api/orders',{currency:'USD',items},{headers:this.headers}).subscribe({next:order=>{this.order.set(order);this.http.delete('/api/cart',{headers:this.headers}).subscribe(()=>this.loadCart());this.busy.set(false)},error:()=>{this.error.set('Checkout failed. Please try again.');this.busy.set(false)}})}
  refreshOrder(){const current=this.order();if(current)this.http.get<Order>(`/api/orders/${current.id}`).subscribe(value=>this.order.set(value))}
}
