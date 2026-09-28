import { CommonModule } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { Component, OnDestroy, OnInit, computed, inject, signal } from '@angular/core';

interface Product { id:string; name:string; description:string; priceCents:number; currency:string; imageUrl:string; inStock:boolean }
interface CartItem { productId:string; quantity:number }
interface Cart { customerId:string; items:CartItem[] }
interface Order { id:string; customerId:string; status:string; currency:string; totalCents:number; items:Array<CartItem & {unitPriceCents:number}> }

type ActivityTone = 'neutral'|'active'|'success'|'danger';
interface ActivityItem { label:string; detail:string; tone:ActivityTone }

@Component({selector:'app-root',imports:[CommonModule],templateUrl:'./app.html',styleUrl:'./app.css'})
export class App implements OnInit, OnDestroy {
  private readonly http=inject(HttpClient);
  private orderPoll?: ReturnType<typeof setInterval>;
  readonly customerId='customer-storefront-demo';

  readonly products=signal<Product[]>([]);
  readonly cart=signal<Cart>({customerId:this.customerId,items:[]});
  readonly order=signal<Order|null>(null);
  readonly loading=signal(true);
  readonly cartBusy=signal(false);
  readonly checkoutBusy=signal(false);
  readonly refreshingOrder=signal(false);
  readonly error=signal('');
  readonly lastOrderRefresh=signal<Date|null>(null);

  readonly cartCount=computed(()=>this.cart().items.reduce((sum,item)=>sum+item.quantity,0));
  readonly cartTotal=computed(()=>this.cart().items.reduce((sum,item)=>sum+item.quantity*(this.product(item.productId)?.priceCents??0),0));
  readonly orderPending=computed(()=>this.order()?.status?.toLowerCase()==='pending');
  readonly activity=computed<ActivityItem[]>(()=>{
    const current=this.order();
    if(!current) return [
      {label:'Waiting for an order',detail:'Place an order to start the distributed workflow.',tone:'neutral'}
    ];
    const status=current.status.toLowerCase();
    if(status==='confirmed') return [
      {label:'Order accepted',detail:'The Order service created the order and published the workflow.',tone:'success'},
      {label:'Event pipeline completed',detail:'Downstream processing completed and the order returned confirmed.',tone:'success'},
      {label:'Customer workflow complete',detail:'The storefront observed the final confirmed state.',tone:'success'}
    ];
    if(status==='cancelled') return [
      {label:'Order accepted',detail:'The Order service created the order.',tone:'success'},
      {label:'Event pipeline failed',detail:'A downstream decision caused the order to be cancelled.',tone:'danger'},
      {label:'Investigation required',detail:'Use service logs and runtime evidence to locate the failing dependency.',tone:'danger'}
    ];
    return [
      {label:'Order accepted',detail:'The Order service created the order.',tone:'success'},
      {label:'Event pipeline in progress',detail:'Kafka-backed downstream processing is still running.',tone:'active'},
      {label:'Waiting for final state',detail:'The UI refreshes the order automatically while it remains pending.',tone:'neutral'}
    ];
  });

  private get headers(){return new HttpHeaders({'X-Customer-ID':this.customerId})}

  ngOnInit(){this.loadProducts();this.loadCart()}
  ngOnDestroy(){this.stopOrderPolling()}

  product(id:string){return this.products().find(value=>value.id===id)}
  money(cents:number){return new Intl.NumberFormat('en-US',{style:'currency',currency:'USD'}).format(cents/100)}

  loadProducts(){
    this.loading.set(true);this.error.set('');
    this.http.get<Product[]>('/api/products').subscribe({
      next:value=>{this.products.set(value);this.loading.set(false)},
      error:()=>{this.error.set('Could not load the catalog. Check the Catalog service and try again.');this.loading.set(false)}
    })
  }

  loadCart(){
    this.http.get<Cart>('/api/cart',{headers:this.headers}).subscribe({
      next:value=>this.cart.set(value),
      error:()=>this.error.set('Could not load your cart. Check the Cart service and Redis dependency.')
    })
  }

  add(product:Product){
    this.cartBusy.set(true);this.error.set('');
    this.http.post('/api/cart/items',{productId:product.id,quantity:1},{headers:this.headers}).subscribe({
      next:()=>{this.loadCart();this.cartBusy.set(false)},
      error:()=>{this.error.set('Could not add the item to the cart.');this.cartBusy.set(false)}
    })
  }

  change(item:CartItem,delta:number){
    const quantity=item.quantity+delta;if(quantity<=0){this.remove(item);return}
    this.cartBusy.set(true);
    this.http.patch(`/api/cart/items/${item.productId}`,{quantity},{headers:this.headers}).subscribe({
      next:()=>{this.loadCart();this.cartBusy.set(false)},
      error:()=>{this.error.set('Could not update the cart item.');this.cartBusy.set(false)}
    })
  }

  remove(item:CartItem){
    this.cartBusy.set(true);
    this.http.delete(`/api/cart/items/${item.productId}`,{headers:this.headers}).subscribe({
      next:()=>{this.loadCart();this.cartBusy.set(false)},
      error:()=>{this.error.set('Could not remove the cart item.');this.cartBusy.set(false)}
    })
  }

  checkout(){
    const items=this.cart().items.map(item=>({...item,unitPriceCents:this.product(item.productId)?.priceCents??0}));
    if(!items.length)return;
    this.checkoutBusy.set(true);this.error.set('');this.stopOrderPolling();
    this.http.post<Order>('/api/orders',{currency:'USD',items},{headers:this.headers}).subscribe({
      next:order=>{
        this.order.set(order);this.lastOrderRefresh.set(new Date());
        this.http.delete('/api/cart',{headers:this.headers}).subscribe({next:()=>this.loadCart()});
        this.checkoutBusy.set(false);
        if(order.status.toLowerCase()==='pending')this.startOrderPolling();
      },
      error:()=>{this.error.set('Checkout failed. Inspect the Order service and its dependencies.');this.checkoutBusy.set(false)}
    })
  }

  refreshOrder(){
    const current=this.order();if(!current||this.refreshingOrder())return;
    this.refreshingOrder.set(true);
    this.http.get<Order>(`/api/orders/${current.id}`).subscribe({
      next:value=>{
        this.order.set(value);this.lastOrderRefresh.set(new Date());this.refreshingOrder.set(false);
        if(value.status.toLowerCase()!=='pending')this.stopOrderPolling();
      },
      error:()=>{this.error.set('Could not refresh the order. The workflow may still be processing.');this.refreshingOrder.set(false)}
    })
  }

  private startOrderPolling(){
    this.stopOrderPolling();
    this.orderPoll=setInterval(()=>this.refreshOrder(),2000);
  }

  private stopOrderPolling(){
    if(this.orderPoll){clearInterval(this.orderPoll);this.orderPoll=undefined}
  }
}
